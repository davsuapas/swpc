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
	"go.uber.org/zap"
)

const ENVConfig = "SW_POOL_CONTROLLER_CONFIG"

const (
	errEnvConfig   = "Environment configuration variable cannot be loaded"
	errLogEncoding = "The log encoding param must be configured to " +
		"('json' or 'console')"
	errLogLevel = "The log level param must be configured to " +
		"(-1: debug, 0: info, 1: Warn, 2: Error, 3: DPanic, 4: Panic, 5: Fatal)"
	errAuthProvider = "The auth provider param must be configured to " +
		"(test, oauth2)"
	errCloudProvider = "The cloud provider param must be configured to " +
		"(none, aws)"
	errDataProvider = "The data provider param must be configured to " +
		"(file, cloud)"
	errHeatbeat = "The heratbeat must be confgured"
	errGets     = "Cannot obtain supplier's secret"
)

const (
	infoApplySecret = "Secret applies"
)

type AuthProvider string

const (
	AuthProviderDev    AuthProvider = "dev"
	AuthProviderOauth2 AuthProvider = "oauth2"
)

type CloudProvider string

const (
	CloudNoProvider  CloudProvider = "none"
	CloudAWSProvider CloudProvider = "aws"
)

type DataProvider string

const (
	FileDataProvider  DataProvider = "file"
	CloudDataProvider DataProvider = "cloud"
)

// Secret manages the secrets
type Secret interface {
	// Get gets the secret in plain text
	Get(secretName string) (map[string]string, error)
}

type DummySecret struct {
}

// Get gets the secret in plain text
// Does not perform any transformation.
func (s *DummySecret) Get(_ string) (map[string]string, error) {
	return make(map[string]string), nil
}

// Server address
type Address struct {
	// TLS defines TLS is used
	TLS bool `json:"tls,omitempty"`
	// Host defines the host
	Host string `json:"host,omitempty"`
	// Port define the port
	Port int `json:"port,omitempty"`
}

// Server describes the http configuration
type Server struct {
	// Internal address on the internal container
	Internal Address `json:"internal,omitempty"`

	// External address on the web
	External Address `json:"external,omitempty"`
}

// InternalURL returns internal URL address
func (s *Server) InternalURL(fragment string) string {
	return url(s.Internal, fragment)
}

// ExternalURL returns external URL address
func (s *Server) ExternalURL(fragment string) string {
	return url(s.External, fragment)
}

// url returns URL address
func url(addr Address, fragment string) string {
	protocol := "http"
	if addr.TLS {
		protocol = "https"
	}

	if addr.Port > 0 {
		return strs.Concat(
			protocol, "://",
			addr.Host,
			":",
			strconv.Itoa(addr.Port),
			"/",
			fragment)
	}

	return strs.Concat(protocol, "://", addr.Host, "/", fragment)
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
	// RedirectURL is the base URL for redirecting provider requests
	// If not defined,
	// it will be formed based on the information in the Server configuration.
	RedirectURL string `json:"redirectProxy,omitempty"`
}

type API struct {
	// CommLatencyTime sets the possible communication latency
	// between the device and the hub, in seconds
	CommLatencyTime int `json:"commLatencyTime,omitempty"`
	// CollectMetricsTime defines how often metrics are collected in miliseconds
	CollectMetricsTime int `json:"collectMetricsTime,omitempty"`
	// ClientID is a identifier that allows the device
	// and the hub to communicate securely.
	ClientID string `json:"clientId,omitempty"`
	// TokenSecretKey defines the secret key to generate the token
	// that allows the device
	// and the hub to communicate securely.
	TokenSecretKey string `json:"tokenSecretKey,omitempty"`
	// HeartbeatInterval is the interval in seconds that
	// the iot device sends a ping for heartbeat
	HeartbeatInterval uint8 `json:"heartbeatInterval"`
	// heartbeatPingTime is the additional time in seconds
	// it may take for the ping to arrive from iot device.
	// Zero does not check heartbeat
	HeartbeatPingTime uint8 `json:"heartbeatPingTime"`
	// HeartbeatTimeoutCount is the amount of timeout allowed
	// before closing the connection to the device.
	HeartbeatTimeoutCount uint8 `json:"heartbeatTimeoutCount"`
}

// Web describes the web configuration
type Web struct {
	// SessionExpiration defines the session expiration in minutes
	SessionExpiration int `json:"expirationSession,omitempty"`
	// SecretKey defines a secret key to AES.
	// It's used in state dance to avoid CRSF.
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

type Cloud struct {
	Provider CloudProvider `json:"provider,omitempty"`
	AWS      AWS           `json:"aws,omitempty"`
}

// AWS defines the global configuration for AWS
type AWS struct {
	// Access key ID
	AKID string `json:"akid,omitempty"`
	// Secret key ID
	SecretKey string `json:"secretKey,omitempty"`
	// Region is the region identifier
	Region string `json:"region,omitempty"`
}

// AWSData defines AWS data configuration
type AWSData struct {
	// TableName is the name table dynamodb
	TableName string `json:"tableName,omitempty"`
}

// FileData defines file data configuration
type FileData struct {
	// FilePath is the data path
	FilePath string `json:"filePath,omitempty"`
}

// Data defines the data configuration
type Data struct {
	// Provider is the data provider
	Provider DataProvider `json:"provider,omitempty"`
	// File is the file data provider
	File FileData `json:"file,omitempty"`
	// AWS is the aws data provider
	AWS AWSData `json:"aws,omitempty"`
}

// Secrets defines the secrets configuration
type Secrets struct {
	// Name is the secrets name
	Name string `json:"name,omitempty"`
}

// Location defines the location info
type Location struct {
	// Location zone
	Zone string `json:"zone,omitempty"`
}

// Config defines the global information
type Config struct {
	Location Location `json:"location,omitempty"`
	Server   `json:"server,omitempty"`
	Zap      `json:"log,omitempty"`
	Web      `json:"web,omitempty"`
	API      `json:"api,omitempty"`
	Hub      `json:"hub,omitempty"`
	Cloud    Cloud   `json:"cloud,omitempty"`
	Secret   Secrets `json:"secret,omitempty"`
	Data     Data    `json:"data,omitempty"`
}

// AuthRedirectURI forms a uri to redirect requests from oauth2 providers
func (c *Config) AuthRedirectURI(fragment string) string {
	fragment = strings.TrimPrefix(fragment, "/")

	if len(c.Auth.RedirectURL) == 0 {
		return c.Server.ExternalURL(fragment)
	}

	return strs.Concat(c.Auth.RedirectURL, "/", fragment)
}

func Default() Config {
	return Config{
		Location: Location{
			Zone: "Europe/Madrid",
		},
		Server: Server{
			Internal: Address{
				TLS:  false,
				Host: "localhost",
				Port: 5000,
			},
			External: Address{
				TLS:  false,
				Host: "localhost",
				Port: 5000,
			},
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
			CommLatencyTime:       2,
			CollectMetricsTime:    1000,
			ClientID:              "sw3kf$fekdy56dfh", // Only for dev
			TokenSecretKey:        "A1Q2wsDE34RF!",    // Only for dev
			HeartbeatInterval:     30,
			HeartbeatPingTime:     5,
			HeartbeatTimeoutCount: 2,
		},
		Hub: Hub{
			TaskTime:         8,
			NotificationTime: 8,
		},
		Cloud: Cloud{
			Provider: CloudNoProvider,
		},
		Data: Data{
			Provider: FileDataProvider,
			File: FileData{
				FilePath: "./data/micro-config.dat",
			},
		},
	}
}

// String returns struct as string
func (c *Config) String() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}

// LoadConfig loads the configuration from environment variable
func LoadConfig() Config { //nolint:cyclop
	cnf := Default()

	env := os.Getenv(ENVConfig)

	if len(env) != 0 {
		if err := json.Unmarshal([]byte(env), &cnf); err != nil {
			panic(
				strs.Format(errEnvConfig, strs.FMTValue("Description", err.Error())))
		}
	}

	if cnf.Encoding != "json" && cnf.Encoding != "console" {
		panic(errLogEncoding)
	}

	if !(cnf.Level >= -1 && cnf.Level <= 5) {
		panic(errLogLevel)
	}

	if cnf.Auth.Provider != AuthProviderDev &&
		cnf.Auth.Provider != AuthProviderOauth2 {
		panic(errAuthProvider)
	}

	if cnf.Cloud.Provider != CloudNoProvider &&
		cnf.Cloud.Provider != CloudAWSProvider {
		panic(errCloudProvider)
	}

	if cnf.Data.Provider != FileDataProvider &&
		cnf.Data.Provider != CloudDataProvider {
		panic(errCloudProvider)
	}

	if cnf.HeartbeatInterval == 0 ||
		cnf.HeartbeatTimeoutCount == 0 {
		panic(errHeatbeat)
	}

	return cnf
}

// ApplySecret calls the secret provider and if the configuration value exists
// contains a secret name preceded by "#",
// applies the value of the secret to the configuration key
func ApplySecret(log *zap.Logger, s Secret, config *Config) {
	secrets, err := s.Get(config.Secret.Name)
	if err != nil {
		panic(strs.Format(errGets, strs.FMTValue("Name", err.Error())))
	}

	if len(secrets) == 0 {
		return
	}

	config.Auth.ClientID = getSecretValue(log, secrets, config.Auth.ClientID)

	config.API.ClientID = getSecretValue(log, secrets, config.API.ClientID)
	config.API.TokenSecretKey = getSecretValue(
		log,
		secrets,
		config.API.TokenSecretKey)

	config.Web.SecretKey = getSecretValue(log, secrets, config.Web.SecretKey)
}

func getSecretValue(
	log *zap.Logger,
	secrets map[string]string,
	value string) string {
	if !strings.HasPrefix(value, "@@") {
		return value
	}

	log.Info(infoApplySecret, zap.String("Secret", value))

	return secrets[value[2:]]
}
