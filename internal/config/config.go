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

const (
	errEnvConfig   = "Environment configuration variable cannot be loaded"
	errLogEncoding = "The log encoding param must be configured to ('json' or 'console')"
	errLogLevel    = "The log level param must be configured to " +
		"(-1: debug, 0: info, 1: Warn, 2: Error, 3: DPanic, 4: Panic, 5: Fatal)"
)

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
	SessionExpiration int `json:"expirationSession,omitempty"`
	// CommLatencyTime sets the possible communication latency between the device and the hub, in seconds
	CommLatencyTime int `json:"commLatencyTime,omitempty"`
	// CheckTransTime defines when the micro-controller is in the time window to transmit in case there are no clients,
	// every so often it checks when to transmit
	CheckTransTime int `json:"checkTransTime,omitempty"`
}

// WebConfig describes the web configuration
type WebConfig struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession,omitempty"`
}

type HubConfig struct {
	// TaskTime defines how often the hub makes maintenance task in seconds
	TaskTime int `json:"taskTime,omitempty"`
	// NotificationTime defines how often a notification is sent in seconds
	NotificationTime int `json:"notificationTime,omitempty"`
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
	HubConfig    `json:"hub,omitempty"`
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
		},
		APIConfig: APIConfig{
			SessionExpiration: 60,
			CommLatencyTime:   5,
			CheckTransTime:    5,
		},
		HubConfig: HubConfig{
			TaskTime:         10,
			NotificationTime: 10,
		},
		DataPath: "./data",
	}

	env := os.Getenv(envConfig)

	if len(env) != 0 {
		if err := json.Unmarshal([]byte(env), &cnf); err != nil {
			panic(strings.Format(errEnvConfig, strings.FMTValue("Description", err.Error())))
		}
	}

	if cnf.Encoding != "json" && cnf.Encoding != "console" {
		panic(errLogEncoding)
	}

	if !(cnf.Level >= -1 && cnf.Level <= 5) {
		panic(errLogLevel)
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
