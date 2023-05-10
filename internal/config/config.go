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
	"strings"

	strs "github.com/swpoolcontroller/pkg/strings"
)

const ENVConfig = "SW_POOL_CONTROLLER_CONFIG"

const (
	errEnvConfig   = "Environment configuration variable cannot be loaded"
	errLogEncoding = "The log encoding param must be configured to ('json' or 'console')"
	errLogLevel    = "The log level param must be configured to " +
		"(-1: debug, 0: info, 1: Warn, 2: Error, 3: DPanic, 4: Panic, 5: Fatal)"
	errAuthProvider = "The auth provider param must be configured to (test, oauth2)"
)

type AuthProvider string

const (
	AuthProviderDev    AuthProvider = "dev"
	AuthProviderOauth2 AuthProvider = "oauth2"
)

// Server describes the http configuration
type Server struct {
	// TLS defines TLS is used
	TLS bool `json:"tls,omitempty"`
	// Host defines the host
	Host string `json:"host,omitempty"`
	// Port define the port
	Port int `json:"port,omitempty"`
}

// URL returns URL address
func (s *Server) URL(fragment string) string {
	protocol := "http"
	if s.TLS {
		protocol = "https"
	}

	if s.Port > 0 {
		return strs.Concat(protocol, "://", s.Host, ":", strconv.Itoa(s.Port), "/", fragment)
	}

	return strs.Concat(protocol, "://", s.Host, "/", fragment)
}

// Auth defines the auth external system based on oauth2
type Auth struct {
	// AuthProvider defines the auth provider. Possible values:
	// test: Only for dev
	// oauth2: Use oauth2
	Provider AuthProvider `json:"provider,omitempty"`
	// ClientID identify the client for oauth2
	ClientID string `json:"clientId,omitempty"`
	// LoginURL defines the page to login
	// The "%redirect_uri" must be added to the signing URL fragment
	// so that the authentication provider knows where to redirect
	// to when the signing process is finished.
	// "%redirect_uri" is filled automatically on the basis of RedirectURL
	// The "%client_id" must be added to the signing URL fragment
	// and is filled with ClientID key
	// The "%state" must be added to the signing URL fragment
	// and is filled with auto generated state to avoid CRSF
	LoginURL string `json:"loginUrl,omitempty"`
	// LogoffURL defines the page to login
	// The "%redirect_uri" must be added to the signing URL fragment
	// so that the authentication provider knows where to redirect
	// to when the signing process is finished.
	// "%redirect_uri" is filled automatically on the basis of RedirectURL
	// The "%client_id" must be added to the signing URL fragment
	// and is filled with ClientID key
	// The "%state" must be added to the signing URL fragment
	// and is filled with auto generated state to avoid CRSF
	LogoutURL string `json:"logoutUrl,omitempty"`
	// JWKURL is the URL to get JWK
	JWKURL string `json:"jwkUrl,omitempty"`
	// TokenURL is the URL to get token
	TokenURL string `json:"tokenUrl,omitempty"`
	// RevokeTokenURL is the URL to rovoke token
	RevokeTokenURL string `json:"revokeTokenUrl,omitempty"`
	// RedirectURL is the base URL for redirecting provider requests
	// If not defined, it will be formed based on the information in the Server configuration.
	RedirectURL string `json:"redirectProxy,omitempty"`
}

type API struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession,omitempty"`
	// CommLatencyTime sets the possible communication latency between the device and the hub, in seconds
	CommLatencyTime int `json:"commLatencyTime,omitempty"`
	// CollectMetricsTime defines how often metrics are collected in miliseconds
	CollectMetricsTime int `json:"collectMetricsTime,omitempty"`
	// CheckTransTime defines when the micro-controller is in the time window to transmit in case there are no clients,
	// every so often it checks when to transmit
	CheckTransTime int `json:"checkTransTime,omitempty"`
	// ClientID is a identifier that allows the device
	// and the hub to communicate securely.
	ClientID string `json:"clientId,omitempty"`
	// SecretKey defines the secret key to generate the token that allows the device
	// and the hub to communicate securely.
	TokenSecretKey string `json:"tokenSecretKey,omitempty"`
}

// Web describes the web configuration
type Web struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession,omitempty"`
	// SecretKey defines a secret key to AES. It's used in state dance to avoid CRSF.
	// Must be of 32 bytes
	SecretKey string `json:"secretKey,omitempty"`
	// Auth defines the auth external system
	Auth Auth `json:"auth,omitempty"`
}

type Hub struct {
	// TaskTime defines how often the hub makes maintenance task in seconds
	TaskTime int `json:"taskTime,omitempty"`
	// NotificationTime defines how often a notification is sent in seconds
	NotificationTime int `json:"notificationTime,omitempty"`
}

// Zap defines the configuration for log framework
type Zap struct {
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
	Server   `json:"server,omitempty"`
	Zap      `json:"log,omitempty"`
	Web      `json:"web,omitempty"`
	API      `json:"api,omitempty"`
	Hub      `json:"hub,omitempty"`
	DataPath string `json:"dataPath,omitempty"`
}

// AuthRedirectURI forms a uri to redirect requests from oauth2 providers
func (c *Config) AuthRedirectURI(fragment string) string {
	fragment = strings.TrimPrefix(fragment, "/")

	if len(c.Auth.RedirectURL) == 0 {
		return c.Server.URL(fragment)
	}

	return strs.Concat(c.Auth.RedirectURL, "/", fragment)
}

func Default() Config {
	return Config{
		Server: Server{
			TLS:  false,
			Host: "localhost",
			Port: 8080,
		},
		Zap: Zap{
			Development: true,
			Level:       -1,
			Encoding:    "console",
		},
		Web: Web{
			SessionExpiration: 10,
			SecretKey:         "123456789asdfghjklzxcvbnmqwertyu", // Only for dev
			Auth: Auth{
				Provider: AuthProviderDev,
			},
		},
		API: API{
			SessionExpiration:  60,
			CommLatencyTime:    2,
			CollectMetricsTime: 800,
			CheckTransTime:     5,
			ClientID:           "sw3kf$fekdy56dfh", // Only for dev
			TokenSecretKey:     "A1Q2wsDE34RF!",    // Only for dev
		},
		Hub: Hub{
			TaskTime:         8,
			NotificationTime: 8,
		},
		DataPath: "./data",
	}
}

// LoadConfig loads the configuration from environment variable
func LoadConfig() Config {
	cnf := Default()

	env := os.Getenv(ENVConfig)

	if len(env) != 0 {
		if err := json.Unmarshal([]byte(env), &cnf); err != nil {
			panic(strs.Format(errEnvConfig, strs.FMTValue("Description", err.Error())))
		}
	}

	if cnf.Encoding != "json" && cnf.Encoding != "console" {
		panic(errLogEncoding)
	}

	if !(cnf.Level >= -1 && cnf.Level <= 5) {
		panic(errLogLevel)
	}

	if cnf.Auth.Provider != AuthProviderDev && cnf.Auth.Provider != AuthProviderOauth2 {
		panic(errAuthProvider)
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
