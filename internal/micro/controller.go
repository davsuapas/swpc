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

package micro

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

const (
	parseStartTimeError = "MicroController.Status-> Parser start time. The micro-controller is set sleep status"
	parseEndTimeError   = "MicroController.Status-> Parser start time. The micro-controller is set sleep status"
)

const layaoutTime = "15:04"

// actions are the actions communication between the device and server
type actions int

const (
	// sleep puts the micro controller to sleep
	sleep actions = iota
	// transmit puts the micro controller to transmit metrics
	transmit
	// checkTransmission puts the micro controller to check transmission status
	checkTransmission
)

type Behavior struct {
	// WakeUpTime is the time set to wake up the micro-controller.
	WakeUpTime uint8
	// CheckTransTime is the time set for the micro to check the status of the clients
	// and whether or not it has to transmit metric
	CheckTransTime uint8
	// Buffer is the buffer in seconds to store metrics int the micro-controller
	Buffer uint8
	Action uint8
}

// String returns struct as string
func (b *Behavior) string() string {
	r, err := json.Marshal(b)
	if err != nil {
		return ""
	}

	return string(r)
}

// Controller controllers the information status on how the micro controller should behave
type Controller struct {
	Log            *zap.Logger
	Hub            *sockets.Hub
	Config         Config
	CheckTransTime uint8

	lock sync.RWMutex
}

// SetConfig updates the micro config into the service
func (c *Controller) SetConfig(conf Config) {
	c.lock.Lock()
	c.Config = conf
	c.lock.Unlock()
}

func (c *Controller) tryConfig() Config {
	var conf Config

	c.lock.RLock()
	conf = c.Config
	c.lock.RUnlock()

	return conf
}

// Download transfers the metrics between micro controller and the hub
// Download also returns the conduct to be taken by the micro-controller
func (c *Controller) Download(metrics string) Behavior {
	c.Hub.Send(metrics)

	actions := c.Actions()

	return actions
}

// Actions gets the conduct to be taken by the micro-controller
func (c *Controller) Actions() Behavior {
	config := c.tryConfig()

	b := Behavior{
		WakeUpTime:     config.Wakeup,
		CheckTransTime: c.CheckTransTime,
		Buffer:         config.Buffer,
		Action:         uint8(c.actions()),
	}

	c.Log.Debug("Micro.Actions", zap.String("Action", b.string()))

	return b
}

// actions gets the actions on how the micro controller should behave
func (c *Controller) actions() actions {
	resp := make(chan sockets.Status)
	c.Hub.Status(resp)
	hstatus := <-resp

	if hstatus == sockets.Closed {
		return sleep
	}

	if hstatus != sockets.Deactivated {
		// All states other than deactivated and closed are susceptible to transmission
		return transmit
	}

	conf := c.tryConfig()

	iniTime, err := time.Parse(layaoutTime, conf.IniSendTime)
	if err != nil {
		c.Log.Fatal(parseStartTimeError, zap.Error(err))

		return sleep
	}

	endTime, err := time.Parse(layaoutTime, conf.EndSendTime)
	if err != nil {
		c.Log.Fatal(parseEndTimeError, zap.Error(err))

		return sleep
	}

	n := time.Now()
	now := time.Date(
		iniTime.Year(),
		iniTime.Month(),
		iniTime.Day(),
		n.Hour(),
		n.Minute(),
		iniTime.Second(),
		iniTime.Nanosecond(),
		iniTime.Location())

	transWindow := false
	// Can transmit within the time zone set by the user.
	if now.After(iniTime) && now.Before(endTime) {
		transWindow = true
	}

	// It is on schedule but there are no clients (the hub is deactivated)
	if transWindow {
		return checkTransmission
	}

	// It is not on schedule and there are no clients (the hub is deactivated)
	return sleep
}
