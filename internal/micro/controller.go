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

package api

import (
	"sync"
	"time"

	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

const (
	parseStartTimeError = "MicroController.Status-> Parser start time. The micro-controller is set sleep status"
	parseEndTimeError   = "MicroController.Status-> Parser start time. The micro-controller is set sleep status"
)

// status are the status communication between the device and server
type Status int

const (
	// Sleep puts the micro controller to sleep. There is nothing to transmit
	Sleep Status = iota
	// Transmit puts the micro controller to transmit metrics
	Transmit
	// Heartbeat puts the micro controller to check socket clients status
	Heartbeat
)

type InfoStatus struct {
	WakeUpTime    uint8
	HeartbeatTime uint8
	Status        Status
}

// Controller controllers the information status on how the micro controller should behave
type Controller struct {
	Log           *zap.Logger
	Hub           *sockets.Hub
	Config        config.MicroConfig
	HeartbeatTime time.Duration

	lock sync.RWMutex
}

// SetConfig updates the micro config into the service
func (s *Controller) SetConfig(conf config.MicroConfig) {
	s.lock.Lock()
	s.Config = conf
	s.lock.Unlock()
}

func (s *Controller) tryConfig() config.MicroConfig {
	var conf config.MicroConfig

	s.lock.RLock()
	conf = s.Config
	s.lock.RUnlock()

	return conf
}

// Status gets the status on how the micro controller should behave
func (s *Controller) Status() Status {
	resp := make(chan sockets.Status)
	s.Hub.Status(resp)
	hstatus := <-resp

	if hstatus == sockets.Closed {
		return Sleep
	}

	if hstatus != sockets.Deactivated {
		// All states other than deactivated and closed are susceptible to transmission
		return Transmit
	}

	conf := s.tryConfig()

	iniTime, err := time.Parse("15:04", conf.IniSendTime)
	if err != nil {
		s.Log.Fatal(parseStartTimeError, zap.Error(err))

		return Sleep
	}

	endTime, err := time.Parse("15:04", conf.EndSendTime)
	if err != nil {
		s.Log.Fatal(parseEndTimeError, zap.Error(err))

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

	transWindow := false
	// Can transmit within the time zone set by the user.
	if now.After(iniTime) && now.Before(endTime) {
		transWindow = true
	}

	// It is on schedule but there are no clients (the hub is deactivated)
	if transWindow {
		return Heartbeat
	}

	// It is not on schedule and there are no clients (the hub is deactivated)
	return Sleep
}
