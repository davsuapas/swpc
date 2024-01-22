/*
 *   Copyright (c) 2022 ELIPCERO
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

package hub

import (
	"github.com/swpoolcontroller/pkg/iot"
	"go.uber.org/zap"
)

const (
	infRegTraces = "Starting the process to register hub traces"
)

// Trace manages the info and errors sent by the hub.
// This info and errors are write into log
type Trace struct {
	log   *zap.Logger
	Trace chan iot.Trace
	Error chan error
}

// NewTrace builds HubError service
func NewTrace(log *zap.Logger) *Trace {
	return &Trace{
		log:   log,
		Trace: make(chan iot.Trace),
		Error: make(chan error),
	}
}

// Run registers errors generated into the hub into the log.
// Launches a gouroutine
func (h *Trace) Register() {
	h.log.Info(infRegTraces)

	go func() {
		for {
			select {
			case e, ok := <-h.Error:
				if !ok {
					return
				}

				h.log.Error("Hub errors", zap.Error(e))
			case t, ok := <-h.Trace:
				if !ok {
					return
				}

				switch t.Level {
				case iot.DebugLevel:
					h.log.Debug(t.Message)
				case iot.InfoLevel:
					h.log.Info(t.Message)
				case iot.WarnLevel:
					h.log.Warn(t.Message)
				case iot.NoneLevel:
				}
			}
		}
	}()
}
