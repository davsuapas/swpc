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

package config

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/swpoolcontroller/pkg/strings"
)

const envConfig = "SW_POOL_CONTROLLER_CONFIG"

// Server describes the http configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// Address returns address to connection server
func (s *ServerConfig) Address() string {
	return strings.Concat(":", strconv.Itoa(s.Port))
}

type APIConfig struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession"`
}

// WebConfig describes the web configuration
type WebConfig struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession"`
	// InactiveCommTime establishes each time the communication between the micro and server is inactive, in seconds
	InactiveCommTime int `json:"inactiveCommTime"`
	// BreakCommTime breakComm parameter establishes each time the communication
	// between the micro and server is break, in seconds
	BreakCommTime int `json:"breakCommTime"`
}

// ZapConfig defines the configuration for log framework
type ZapConfig struct {
	// Development mode. Common value: false
	Development bool `json:"development,omitempty"`
	// Level. See logging.Level. Common value: Depending of Development flag
	Level int `json:"level,omitempty"`
	// Encoding type. Common value: Depending of Development flag
	// The values can be: j -> json format, c -> console format
	Encoding string `json:"encoding,omitempty"`
}

// Config defines the global information
type Config struct {
	ServerConfig `json:"server,omitempty"`
	ZapConfig    `json:"log,omitempty"`
	WebConfig    `json:"web,omitempty"`
	APIConfig    `json:"api,omitempty"`
	DataPath     string `json:"dataPath,omitempty"`
}

// LoadConfig loads the configuration from environment variable
func LoadConfig() Config {
	cnf := Config{
		ServerConfig: ServerConfig{Port: 8080},
		ZapConfig: ZapConfig{
			Development: true,
			Level:       -1,
			Encoding:    "console",
		},
		WebConfig: WebConfig{
			SessionExpiration: 10,
			InactiveCommTime:  10,
			BreakCommTime:     40,
		},
		APIConfig: APIConfig{
			SessionExpiration: 60,
		},
		DataPath: "./data",
	}

	env := os.Getenv(envConfig)

	if len(env) != 0 {
		if err := json.Unmarshal([]byte(env), &cnf); err != nil {
			panic(strings.Concat("Configuration environment variable cannot be loaded. Description ", err.Error()))
		}
	}

	if cnf.Encoding != "json" && cnf.Encoding != "console" {
		panic("The log encoding param must be configured to ('json' or 'console')")
	}

	if !(cnf.Level >= -1 && cnf.Level <= 5) {
		panic("The log level param must be configured to " +
			"(-1: debug, 0: info, 1: Warn, 2: Error, 3: DPanic, 4: Panic, 5: Fatal)")
	}

	if cnf.InactiveCommTime > cnf.BreakCommTime {
		panic("The InactiveCommTime cannot be greater than BreakCommTime")
	}

	return cnf
}

// String returns struct as string
func (c *Config) String() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}
