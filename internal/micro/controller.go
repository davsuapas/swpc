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
	"strconv"
	"sync"
	"time"

	"github.com/swpoolcontroller/pkg/sockets"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const (
	parseStartTimeError = "MicroController.Status-> Parser start time. The micro-controller is set sleep status"
	parseEndTimeError   = "MicroController.Status-> Parser end time. The micro-controller is set sleep status"
)

const layaoutTime = "15:04"

// Actions are the Actions communication between the device and server
type Actions int

const (
	// Sleep puts the micro controller to Sleep
	Sleep Actions = iota
	// Transmit puts the micro controller to Transmit metrics
	Transmit
	// CheckTransmission puts the micro controller to check transmission status
	CheckTransmission
)

type Behavior struct {
	// WakeUpTime is the time set to wake up the micro-controller.
	WakeUpTime uint8
	// CheckTransTime is the time set for the micro to check the status of the clients
	// and whether or not it has to transmit metric
	CheckTransTime uint8
	// CollectMetricsTime defines how often metrics are collected
	CollectMetricsTime uint16
	// Buffer is the buffer in seconds to store metrics int the micro-controller
	Buffer uint8
	Action uint8
}

// String returns struct as string
func (b *Behavior) String() string {
	return strings.Concat("WakeUpTime: ", string(b.WakeUpTime),
		", CheckTransTime: ", string(b.CheckTransTime),
		", CollectMetricsTime: ", strconv.Itoa(int(b.CollectMetricsTime)),
		", Buffer: ", string(b.Buffer),
		", Action: ", string(b.Action))
}

// Hub manages the socket pool and distribute messages
type Hub interface {
	// Send sends message to the all clients into hub
	Send(message string)
	// Status request hub status via channel
	Status(resp chan sockets.Status)
}

// Controller controllers the information status on how the micro controller should behave
type Controller struct {
	Log                *zap.Logger
	Hub                Hub
	Config             Config
	CheckTransTime     uint8
	CollectMetricsTime uint16

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

	return Behavior{
		WakeUpTime:         config.Wakeup,
		CheckTransTime:     c.CheckTransTime,
		CollectMetricsTime: c.CollectMetricsTime,
		Buffer:             config.Buffer,
		Action:             uint8(c.actions()),
	}
}

// actions gets the actions on how the micro controller should behave
func (c *Controller) actions() Actions {
	resp := make(chan sockets.Status)
	c.Hub.Status(resp)
	hstatus := <-resp

	if hstatus == sockets.Closed {
		return Sleep
	}

	if hstatus != sockets.Deactivated {
		// All states other than deactivated and closed are susceptible to transmission
		return Transmit
	}

	conf := c.tryConfig()

	iniTime, err := time.Parse(layaoutTime, conf.IniSendTime)
	if err != nil {
		c.Log.Error(parseStartTimeError, zap.Error(err))

		return Sleep
	}

	endTime, err := time.Parse(layaoutTime, conf.EndSendTime)
	if err != nil {
		c.Log.Error(parseEndTimeError, zap.Error(err))

		return Sleep
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

	// Can transmit within the time zone set by the user.
	// It is on schedule but there are no clients (the hub is deactivated)
	if now.After(iniTime) && now.Before(endTime) {
		return CheckTransmission
	}

	// It is not on schedule and there are no clients (the hub is deactivated)
	return Sleep
}
