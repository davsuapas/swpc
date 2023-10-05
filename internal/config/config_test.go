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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/config/mocks"
)

func TestLoadConfig(t *testing.T) {
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
			env: `{
				"server": {
					"Internal": {
						"tls": true,
						"host": "192.168.100.1",
						"port": 2020
					},
					"External": {
						"tls": true,
						"host": "192.168.100.2",
						"port": 2021
					}
				},
				"log": {"development": false, "level": 3},
				"web": {
					"expirationSession": 15,
					"secretKey": "123",
					"auth": {
						"provider": "oauth2",
						"clientId": "clientId",
						"loginUrl": "loginUrl",
						"jwkUrl": "jwkUrl",
						"tokenUrl": "tokenUrl",
						"redirectLoginUri": "redirectLoginUri",
						"redirectLogoutUri": "redirectLogoutUri"
					}
				},
				"api": {
					"expirationSession": 15,
					"commLatencyTime": 4,
					"collectMetricsTime": 10,
					"checkTransTime": 4,
					"clientId": "123",
					"tokenSecretKey": "123"
				},
				"hub": {"taskTime": 6, "notificationTime": 7},
				"cloud": {
					"provider": "aws",
					"aws": {
						"akid": "akid",
						"secretKey": "secretKey",
						"secret": {
							"region": "region",
							"name": "name"
						}
					}
				},
				"dataPath": "./datas"
			}`,
			res: config.Config{
				Server: config.Server{
					Internal: config.Address{
						TLS:  true,
						Host: "192.168.100.1",
						Port: 2020,
					},
					External: config.Address{
						TLS:  true,
						Host: "192.168.100.2",
						Port: 2021,
					},
				},
				Zap: config.Zap{
					Development: false,
					Level:       3,
					Encoding:    "console",
				},
				Web: config.Web{
					SessionExpiration: 15,
					SecretKey:         "123",
					Auth: config.Auth{
						Provider: "oauth2",
						ClientID: "clientId",
						LoginURL: "loginUrl",
						JWKURL:   "jwkUrl",
						TokenURL: "tokenUrl",
					},
				},
				API: config.API{
					SessionExpiration:  15,
					CommLatencyTime:    4,
					CollectMetricsTime: 10,
					CheckTransTime:     4,
					ClientID:           "123",
					TokenSecretKey:     "123",
				},
				Hub: config.Hub{
					TaskTime:         6,
					NotificationTime: 7,
				},
				Cloud: config.Cloud{
					Provider: config.CloudAWSProvider,
					AWS: config.AWS{
						AKID:      "akid",
						SecretKey: "secretKey",
						Secret: config.AWSSecret{
							Region: "region",
							Name:   "name",
						},
					},
				},
				DataPath: "./datas",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(config.ENVConfig, tt.env)

			assert.Equal(t, tt.res, config.LoadConfig())
		})
	}
}

func TestLoadConfig_Panic(t *testing.T) {
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
			env:  `{"log": {"encoding": "no_exist"}}`,
		},
		{
			name: "Config. Level incorrect",
			env:  `{"log": {"level": 8}}`,
		},
		{
			name: "Config. Auth provider incorrect",
			env:  `{"web": {"auth": { "provider": "no_exist"}}}`,
		},
		{
			name: "Config. Cloud provider incorrect",
			env:  `{"cloud": { "provider": "no_exist"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(config.ENVConfig, tt.env)

			assert.Panics(t, func() { config.LoadConfig() })
		})
	}
}

func TestServer_InternalURL(t *testing.T) {
	type fields struct {
		TSL  bool
		Port int
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "URL with TLS",
			fields: fields{
				TSL:  true,
				Port: 0,
			},
			want: "https://localhost/fragment",
		},
		{
			name: "URL without TSL",
			fields: fields{
				TSL:  false,
				Port: 2020,
			},
			want: "http://localhost:2020/fragment",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := config.Server{
				Internal: config.Address{
					TLS:  tt.fields.TSL,
					Host: "localhost",
					Port: tt.fields.Port,
				},
			}
			res := s.InternalURL("fragment")

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestServer_ExternalURL(t *testing.T) {
	s := config.Server{
		External: config.Address{
			TLS:  true,
			Host: "localhost",
			Port: 2020,
		},
	}
	res := s.ExternalURL("fragment")

	assert.Equal(t, "https://localhost:2020/fragment", res)
}

func TestConfig_AuthRedirectURI(t *testing.T) {
	type fields struct {
		RedirectURL string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "AuthRedirectURI with redirect URL",
			fields: fields{
				RedirectURL: "",
			},
			want: "http://server:2020/fragment",
		},
		{
			name: "AuthRedirectURI without redirect URL",
			fields: fields{
				RedirectURL: "http://server",
			},
			want: "http://server/fragment",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Config{
				Server: config.Server{
					External: config.Address{
						TLS:  false,
						Host: "server",
						Port: 2020,
					},
				},
				Web: config.Web{
					Auth: config.Auth{
						RedirectURL: tt.fields.RedirectURL,
					},
				},
			}

			res := c.AuthRedirectURI("fragment")

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestServerConfig_String(t *testing.T) {
	t.Parallel()

	c := config.Default()

	res := c.String()

	assert.NotEmpty(t, res)
}

func TestApplySecret(t *testing.T) {
	t.Parallel()

	type args struct {
		config  config.Config
		secrets map[string]string
	}

	tests := []struct {
		name     string
		args     args
		expected config.Config
	}{
		{
			name: "Apply Secret. Secret name not exists. It should return the same configuration values",
			args: args{
				config: config.Config{
					Web: config.Web{
						SecretKey: "@@SecretKey",
						Auth: config.Auth{
							ClientID: "@@ClientID",
						},
					},
					API: config.API{
						ClientID:       "@@ClientID",
						TokenSecretKey: "@@TokenSecretKey",
					},
					DataPath: "",
				},
				secrets: map[string]string{},
			},
			expected: config.Config{
				Web: config.Web{
					SecretKey: "@@SecretKey",
					Auth: config.Auth{
						ClientID: "@@ClientID",
					},
				},
				API: config.API{
					ClientID:       "@@ClientID",
					TokenSecretKey: "@@TokenSecretKey",
				},
				DataPath: "",
			},
		},
		{
			name: "Apply Secret. Secret name exists. It should return the configuration values applying secrets",
			args: args{
				config: config.Config{
					Web: config.Web{
						SecretKey: "@@SecretKey",
						Auth: config.Auth{
							ClientID: "@@ClientID",
						},
					},
					API: config.API{
						ClientID:       "ClientID",
						TokenSecretKey: "@@TokenSecretKey",
					},
				},
				secrets: map[string]string{
					"SecretKey":      "123",
					"TokenSecretKey": "1234",
				},
			},
			expected: config.Config{
				Web: config.Web{
					SecretKey: "123",
					Auth: config.Auth{
						ClientID: "",
					},
				},
				API: config.API{
					ClientID:       "ClientID",
					TokenSecretKey: "1234",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewSecret(t)
			s.On("Get", tt.args.config.Cloud.AWS.Secret.Name).Return(tt.args.secrets, nil)

			config.ApplySecret(s, &tt.args.config)

			assert.Equal(t, tt.expected, tt.args.config)

			s.AssertExpectations(t)
		})
	}
}
