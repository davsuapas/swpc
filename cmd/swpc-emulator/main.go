/*
 *   Copyright (c) 2022 CARISA
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

// Simulates the process on an arduino board,
// in charge of collecting metrics from a pool to be sent in real time to a server.
// The code is thought as if it was for the C language.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/swpoolcontroller/pkg/strings"
)

const (
	URL              = "http://localhost:8080"
	URLAUTH          = URL + "/auth"
	URLAPI           = URL + "/micro/api"
	sID              = "sw3kf$fekdy56dfh"
	retryTimeSeconds = 10
)

const (
	// sleep puts the micro controller to sleep.
	sleepServer = 0
	// transmit puts the micro controller to transmit metrics
	transmitServer = 1
	// checkTransmission puts the micro controller to check transmission status
	checkTransmissionServer = 2
)

// Actions of the states machine
const (
	sleep               = 0
	transmit            = 1
	checkNextAction     = 2
	waitCheckNextAction = 3
	waitRetry           = 4
	waitBuffer          = 5
)

// See micro.Behavior
type Config struct {
	WakeUpTime         uint8
	CheckTransTime     uint8
	CollectMetricsTime uint16
	Buffer             uint8
	Action             uint8
}

type Retry struct {
	action  uint8
	counter uint8
	max     uint8
}

type Buffer struct {
	bufferTemp      string
	bufferPH        string
	bufferCL        string
	timeStartChrono int64
	nextAction      uint8
}

var (
	config          Config
	retry           Retry
	buffer          Buffer
	action          uint8
	securToken      string
	timeStartChrono int64
	lastError       string
)

func main() {
	setup()

	for {
		loop()
		time.Sleep(1 * time.Millisecond)
	}
}

func setup() {
	config = Config{
		WakeUpTime:         30,
		CheckTransTime:     5,
		CollectMetricsTime: 800,
		Buffer:             3,
	}

	collectBuffer(checkNextAction)
}

func loop() {
	switch action {
	case sleep:
		fmt.Println("sleeping ...")
		time.Sleep(time.Duration(int64(config.WakeUpTime)) * time.Minute)
		collectBuffer(checkNextAction)
	case checkNextAction:
		queryNextAction(checkNextAction)
	case transmit:
		if doTrasnmisssion() {
			// doTransmission set action
			collectBuffer(action)
		}
	case waitBuffer:
		wakeupBuffer()
	case waitRetry:
		wakeupRetry()
	case waitCheckNextAction:
		wakeupCheckNextAction()
	}
}

// queryNextAction checks against the server the new actions to be performed on the micro-controller
// If there is error retry previous action. Also it set next action
func queryNextAction(actionRetry uint8) {
	fmt.Println("queryNextAction")

	do(http.MethodGet, "/action", "", actionRetry, 30)
}

// doTrasnmisssion trasmits sensor buffer and set next action
func doTrasnmisssion() bool {
	fmt.Println("doTrasnmisssion")

	buffer := strings.Concat(buffer.bufferTemp, ";", buffer.bufferPH, ";", buffer.bufferCL)

	return do(http.MethodPost, "/download", buffer, transmit, 60)
}

// do makes http server request, also controls failures
func do(method string, uri string, body string, actionRetry uint8, maxRetry uint8) bool {
	var res string

	if !REST(method, strings.Concat(URLAPI, uri), body, &res) {
		retryAction(actionRetry, maxRetry)

		return false
	}

	if err := json.Unmarshal([]byte(res), &config); err != nil {
		setLastError(err.Error())

		retryAction(actionRetry, maxRetry)

		return false
	}

	fmt.Println("do.config: ", res)

	iniRetry()

	setNextAction(config.Action)

	return true
}

// wakeupBuffer collects metrics each seconds
// if it reaches the maximum buffer time, it performs the action indicated in config.Buffer
func wakeupBuffer() {
	timeElapsedMetricsSec := (time.Now().UnixMilli() - buffer.timeStartChrono) / 1000

	if timeElapsedMetricsSec > int64(config.Buffer) {
		fmt.Println(
			"wakeupBuffer.timeElapsedMetricsSec: ", timeElapsedMetricsSec,
			", nextAction: ", printAction(buffer.nextAction))

		action = buffer.nextAction

		return
	}

	if timeElapsedMiliSec() >= int64(config.CollectMetricsTime) {
		startChrono()
		collectMetrics()
	}
}

// wakeupRetry checks the retry time. If it is met, the action is reattempted,
// and if the maximum time is met, the action is slept.
func wakeupRetry() {
	if timeElapsedSec() >= retryTimeSeconds {
		if retry.counter == retry.max {
			iniRetry()

			fmt.Println("wakeupRetry.retry-max: ", retry.max)

			action = sleep

			return
		}

		retry.counter++

		fmt.Println("wakeupRetry.counter: ", retry.counter, ", action: ", printAction(retry.action))

		action = retry.action
	}
}

// wakeupCheckNextAction checks if he has to wake up to get next action
func wakeupCheckNextAction() {
	if timeElapsedSec() > int64(config.CheckTransTime) {
		fmt.Println("wakeupCheckNextAction.timeElapsedSec(): ", timeElapsedSec())

		action = checkNextAction
	}
}

// setNextAction calculates the next actions
func setNextAction(actionServer uint8) {
	switch actionServer {
	case sleepServer:
		action = sleep
	case transmitServer:
		action = transmit
	case checkTransmissionServer:
		action = waitCheckNextAction

		startChrono()
	}

	fmt.Println("nextAction: ", printAction(action))
}

// collectBuffer requests to perform the action of collecting and it set next action by nextAction param
func collectBuffer(nextAction uint8) {
	fmt.Println("collectBuffer.nextAction: ", printAction(nextAction))

	buffer.timeStartChrono = time.Now().UnixMilli()
	buffer.nextAction = nextAction
	buffer.bufferTemp = ""
	buffer.bufferPH = ""
	buffer.bufferCL = ""

	action = waitBuffer

	startChrono()
}

// collectMetrics collects metrics of the sensor
func collectMetrics() {
	fmt.Println("collectMetrics")

	if len(buffer.bufferTemp) > 0 {
		buffer.bufferTemp = strings.Concat(buffer.bufferTemp, ",")
	}

	buffer.bufferTemp = strings.Concat(buffer.bufferTemp, strconv.Itoa(tempSensor()))

	if len(buffer.bufferPH) > 0 {
		buffer.bufferPH = strings.Concat(buffer.bufferPH, ",")
	}

	buffer.bufferPH = strings.Concat(buffer.bufferPH, strconv.Itoa(phSensor()))

	if len(buffer.bufferCL) > 0 {
		buffer.bufferCL = strings.Concat(buffer.bufferCL, ",")
	}

	buffer.bufferCL = strings.Concat(buffer.bufferCL, fmt.Sprintf("%f", clSensor()))
}

func tempSensor() int {
	return rand.Intn(40)
}

func phSensor() int {
	return rand.Intn(14)
}

func clSensor() float32 {
	var min float32 = 3

	var max float32 = 8.5

	return rand.Float32()*(max-min) + min
}

// REST Http Client. it permits oauth by token
func REST(method string, url string, body string, result *string) bool {
	fmt.Println("REST.method: ", method, ", URL: ", url, "Body: ", body)

	var res string

	if !restInternal(true, true, method, url, body, &res) {
		return false
	}

	*result = res

	return true
}

func restInternal(
	auth bool,
	checkStatusUnauthorized bool,
	method string,
	url string,
	body string,
	result *string) bool {
	// If there is no security token, it requests it to be added to the request header.
	if auth {
		oAuthToken()
	}

	ctx := context.TODO()

	bodyb := bytes.NewBuffer([]byte{})

	if len(body) > 0 {
		bodyb = bytes.NewBuffer([]byte(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyb)
	if err != nil {
		setLastError(err.Error())

		return false
	}

	req.Header.Set("Content-Type", "text/plain")

	if auth {
		req.Header.Add("Authorization", strings.Concat("Bearer ", securToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		setLastError(err.Error())

		return false
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		setLastError(err.Error())

		return false
	}

	// StatusUnauthorized means that the token has to be renewed.
	// and re-launching the request
	if checkStatusUnauthorized && resp.StatusCode == http.StatusUnauthorized {
		securToken = ""

		var res string

		if errb := restInternal(auth, false, method, url, body, &res); !errb {
			return false
		}

		*result = res

		return true
	}

	if resp.StatusCode == http.StatusOK {
		*result = string(data)

		return true
	}

	return false
}

// oAuthToken gets token of security
// If there is no security token, it requests it to be added to the request header
func oAuthToken() bool {
	if len(securToken) == 0 {
		var res string

		if err := restInternal(false, false, http.MethodGet, strings.Concat(URLAUTH, "/token/", sID), "", &res); !err {
			return false
		}

		securToken = res

		fmt.Println("oAuthToken.securToken: ", securToken)
	}

	return true
}

// iniRetry initilizes counter
func iniRetry() {
	retry.counter = 0
}

// retryAction requests retry activating chrone
func retryAction(raction uint8, max uint8) {
	fmt.Println("retryAction.action: ", raction, ", max: ", max)

	retry.action = raction
	retry.max = max

	action = waitRetry

	startChrono()
}

func startChrono() {
	timeStartChrono = time.Now().UnixMilli()
}

func timeElapsedMiliSec() int64 {
	return time.Now().UnixMilli() - timeStartChrono
}

func timeElapsedSec() int64 {
	return timeElapsedMiliSec() / 1000
}

func setLastError(le string) {
	lastError = le
	fmt.Println("lastError: ", lastError)
}

func printAction(a uint8) string {
	switch a {
	case sleep:
		return "sleep"
	case transmit:
		return "transmit"
	case checkNextAction:
		return "checkNextAction"
	case waitCheckNextAction:
		return "waitCheckNextAction"
	case waitRetry:
		return "waitRetry"
	case waitBuffer:
		return "waitBuffer"
	}

	return "None"
}
