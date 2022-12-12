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

package internal

import "go.uber.org/zap"

// HubTrace manages the info and errors sent by the hub. This info and errors are write into log
type HubTrace struct {
	log    *zap.Logger
	Infos  chan []string
	Errors chan []error
}

// NewHubTrace builds HubError service
func NewHubTrace(log *zap.Logger) *HubTrace {
	return &HubTrace{
		log:    log,
		Infos:  make(chan []string),
		Errors: make(chan []error),
	}
}

// Run registers errors generated into the hub into the log. Launches a gouroutine
func (h *HubTrace) Register() {
	h.log.Info("Starting the process to register hub traces")

	go func() {
		for {
			select {
			case errors := <-h.Errors:
				for _, e := range errors {
					h.log.Error("Hub errors", zap.Error(e))
				}
			case infos := <-h.Infos:
				for _, i := range infos {
					h.log.Info(i)
				}
			}
		}
	}()
}