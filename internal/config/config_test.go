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

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  string
		res  config.Config
	}{
		{
			name: "Config. Default",
			env:  "",
			res:  config.Default(),
		},
		{
			name: "Config. Custom",
			env:  `{"server": {"port": 2020}}`,
			res: config.Config{
				ServerConfig: config.ServerConfig{
					Port: 2020,
				},
				ZapConfig: config.ZapConfig{
					Development: true,
					Level:       -1,
					Encoding:    "console",
				},
				WebConfig: config.WebConfig{
					SessionExpiration: 10,
				},
				APIConfig: config.APIConfig{
					SessionExpiration:  60,
					CommLatencyTime:    2,
					CollectMetricsTime: 800,
					CheckTransTime:     5,
				},
				HubConfig: config.HubConfig{
					TaskTime:         8,
					NotificationTime: 8,
				},
				DataPath: "./data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(config.ENVConfig, tt.env)

			cnf := config.LoadConfig()

			assert.Equal(t, tt.res, cnf)
		})
	}
}

func TestLoadConfig_Panic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  string
	}{
		{
			name: "Config. Unmarshal error",
			env:  "{",
		},
		{
			name: "Config. Encoding incorrect",
			env:  `{"log": {"encoding": "other"}}`,
		},
		{
			name: "Config. Level incorrect",
			env:  `{"log": {"level": 8}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(config.ENVConfig, tt.env)

			assert.Panics(t, func() { config.LoadConfig() })
		})
	}
}

func TestServerConfig_String(t *testing.T) {
	c := config.Default()

	res := c.String()

	assert.NotEmpty(t, res)
}

func TestServerConfig_Address(t *testing.T) {
	s := config.ServerConfig{
		Port: 2020,
	}

	res := s.Address()

	assert.Equal(t, ":2020", res)
}
