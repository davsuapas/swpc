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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/micro/mocks"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

func TestConfigRead_Read(t *testing.T) {
	t.Parallel()

	type errors struct {
		want bool
		msg  string
	}

	tests := []struct {
		name     string
		dataPath string
		expected micro.Config
		err      errors
	}{
		{
			name:     "Read config from file",
			dataPath: "./testr/micro-config.dat",
			expected: micro.Config{
				IniSendTime: "10:00",
				EndSendTime: "21:01",
				Wakeup:      16,
				Buffer:      3,
			},
			err: errors{
				want: false,
			},
		},
		{
			name:     "Read config when the file don't exist",
			dataPath: "./testr/micro-no_exist.dat",
			expected: micro.DefaultConfig(),
			err: errors{
				want: false,
			},
		},
		{
			name:     "Read config. Error unserialize",
			dataPath: "./testr/micro-config-error.dat",
			expected: micro.DefaultConfig(),
			err: errors{
				want: true,
				msg:  "Unmarshalling the configuration",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := micro.ConfigRead{
				Log:      zap.NewExample(),
				DataFile: tt.dataPath,
			}

			got, err := c.Read()
			if err != nil && tt.err.want {
				assert.ErrorContains(t, err, tt.err.msg, "Error")
			}

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConfigWrite_Save(t *testing.T) {
	t.Parallel()

	type fields struct {
		DataFile string
	}

	type args struct {
		data micro.Config
	}

	type res struct {
		scnf sockets.Config
	}

	type errors struct {
		want bool
		msg  string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
		err    errors
	}{
		{
			name: "Write micro config successfully",
			fields: fields{
				DataFile: "./testr/micro-config-write-sucess.dat",
			},
			args: args{
				data: micro.DefaultConfig(),
			},
			res: res{
				scnf: sockets.Config{
					CommLatency:      2 * time.Second,
					Buffer:           3 * time.Second,
					TaskTime:         8 * time.Second,
					NotificationTime: 8 * time.Second,
				},
			},
			err: errors{
				want: false,
			},
		},
		{
			name: "Write micro config. Error writing file",
			fields: fields{
				DataFile: "./no_exist/micro-config-write-sucess.dat",
			},
			args: args{
				data: micro.DefaultConfig(),
			},
			err: errors{
				want: true,
				msg:  "Saving the configuration for the micro controller",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cnf := config.Default()

			h := mocks.NewHub(t)

			if !tt.err.want {
				h.On("Config", tt.res.scnf)
			}

			c := micro.ConfigWrite{
				Log:      zap.NewExample(),
				MControl: &micro.Controller{},
				Hub:      h,
				Config:   cnf,
				DataFile: tt.fields.DataFile,
			}

			if err := c.Save(tt.args.data); err != nil && tt.err.want {
				assert.ErrorContains(t, err, tt.err.msg, "Error")

				return
			}

			if err := os.Remove(tt.fields.DataFile); err != nil {
				assert.Error(t, err, "Removing micro config file created")
			}

			h.AssertExpectations(t)
		})
	}
}
