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

package micro_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/micro/mocks"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

func TestController_SetConfig(t *testing.T) {
	t.Parallel()

	cnf := micro.ConfigDefault()
	c := &micro.Controller{}

	c.SetConfig(cnf)

	assert.Equal(t, cnf, c.Config)
}

func TestController_Download(t *testing.T) {
	t.Parallel()

	mockHub := mocks.NewHub(t)

	metrics := "1,2,3"

	mockHub.On("Send", metrics)
	mockHub.On("Status", mock.AnythingOfType("chan sockets.Status")).Run(func(args mock.Arguments) {
		if s, ok := args.Get(0).(chan sockets.Status); ok {
			go func() {
				s <- sockets.Closed
			}()
		}
	})

	cnf := micro.ConfigDefault()

	c := &micro.Controller{
		Log:                zap.NewExample(),
		Hub:                mockHub,
		Config:             cnf,
		CheckTransTime:     1,
		CollectMetricsTime: 2,
	}

	b := c.Download(metrics)

	assert.Equal(
		t,
		micro.Behavior{
			WakeUpTime:         cnf.Wakeup,
			CheckTransTime:     c.CheckTransTime,
			CollectMetricsTime: c.CollectMetricsTime,
			Buffer:             cnf.Buffer,
			Action:             uint8(micro.Sleep),
		},
		b)

	mockHub.AssertExpectations(t)
}
