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

func TestBehavior_String(t *testing.T) {
	t.Parallel()

	b := micro.Behavior{
		WakeUpTime:         1,
		CheckTransTime:     1,
		CollectMetricsTime: 1,
		Buffer:             1,
		Action:             1,
	}

	s := b.String()

	assert.Equal(t, "WakeUpTime: \x01, CheckTransTime: \x01, CollectMetricsTime: 1, Buffer: \x01, Action: \x01", s)
}

func TestController_SetConfig(t *testing.T) {
	t.Parallel()

	cnf := micro.DefaultConfig()
	c := &micro.Controller{}

	c.SetConfig(cnf)

	assert.Equal(t, cnf, c.Config)
}

func TestController_Download(t *testing.T) {
	t.Parallel()

	metrics := "1,2,3"

	type fields struct {
		cnf    micro.Config
		status sockets.Status
	}

	tests := []struct {
		name   string
		fields fields
		res    micro.Behavior
	}{
		{
			name: "Download. Socket closed",
			fields: fields{
				cnf:    micro.DefaultConfig(),
				status: sockets.Closed,
			},
			res: micro.Behavior{
				WakeUpTime:         30,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
				Buffer:             3,
				Action:             uint8(micro.Sleep),
			},
		},
		{
			name: "Download. Socket unlike Deactivated",
			fields: fields{
				cnf:    micro.DefaultConfig(),
				status: sockets.Streaming,
			},
			res: micro.Behavior{
				WakeUpTime:         30,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
				Buffer:             3,
				Action:             uint8(micro.Transmit),
			},
		},
		{
			name: "Download. Bad formatted start time",
			fields: fields{
				cnf: micro.Config{
					IniSendTime: "123-221",
					EndSendTime: "",
					Wakeup:      20,
					Buffer:      5,
				},
				status: sockets.Deactivated,
			},
			res: micro.Behavior{
				WakeUpTime:         20,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
				Buffer:             5,
				Action:             uint8(micro.Sleep),
			},
		},
		{
			name: "Download. Bad formatted end time",
			fields: fields{
				cnf: micro.Config{
					IniSendTime: "12:12",
					EndSendTime: "123-221",
					Wakeup:      20,
					Buffer:      5,
				},
				status: sockets.Deactivated,
			},
			res: micro.Behavior{
				WakeUpTime:         20,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
				Buffer:             5,
				Action:             uint8(micro.Sleep),
			},
		},
		{
			name: "Download. Socket deactivated",
			fields: fields{
				cnf: micro.Config{
					IniSendTime: "11:52",
					EndSendTime: "11:52",
					Wakeup:      20,
					Buffer:      5,
				},
				status: sockets.Deactivated,
			},
			res: micro.Behavior{
				WakeUpTime:         20,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
				Buffer:             5,
				Action:             uint8(micro.Sleep),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mhub := mocks.NewHub(t)

			mhub.On("Send", metrics)
			mhub.On("Status", mock.AnythingOfType("chan sockets.Status")).Run(func(args mock.Arguments) {
				if s, ok := args.Get(0).(chan sockets.Status); ok {
					go func() {
						s <- tt.fields.status
					}()
				}
			})

			c := &micro.Controller{
				Log:                zap.NewExample(),
				Hub:                mhub,
				Config:             tt.fields.cnf,
				CheckTransTime:     1,
				CollectMetricsTime: 2,
			}

			b := c.Download(metrics)

			assert.Equal(t, tt.res, b)

			mhub.AssertExpectations(t)
		})
	}
}
