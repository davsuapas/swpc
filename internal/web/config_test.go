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

package web_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/micro/mocks"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

func TestConfigWeb_Load(t *testing.T) {
	t.Parallel()

	zap := zap.NewExample()

	type res struct {
		status int
		body   string
	}

	tests := []struct {
		name     string
		dataFile string
		res
	}{
		{
			name:     "Load. StatusOk",
			dataFile: "./testr/micro-config.dat",
			res: res{
				status: http.StatusOK,
				body:   "{\"iniSendTime\":\"10:00\",\"endSendTime\":\"21:01\",\"wakeup\":16,\"buffer\":3}\n",
			},
		},
		{
			name:     "Load. StatusInternalServerError",
			dataFile: "./testr/micro-config-error.dat",
			res: res{
				status: http.StatusInternalServerError,
				body:   "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/config", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			cf := &web.ConfigWeb{
				Log: zap,
				MicroR: &micro.ConfigRead{
					Log:      zap,
					DataFile: tt.dataFile,
				},
				MicroW: &micro.ConfigWrite{},
			}

			_ = cf.Load(c)

			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, tt.body, rec.Body.String())
		})
	}
}

func TestConfigWeb_Save(t *testing.T) {
	t.Parallel()

	zap := zap.NewExample()

	type fields struct {
		hubf     func() micro.Hub
		dataFile string
	}

	tests := []struct {
		name      string
		field     fields
		argBody   string
		resStatus int
	}{
		{
			name: "Save. StatusOk",
			field: fields{
				hubf: func() micro.Hub {
					h := mocks.NewHub(t)
					h.On("Config", sockets.Config{
						CommLatency:      2 * time.Second,
						Buffer:           10 * time.Second,
						TaskTime:         8 * time.Second,
						NotificationTime: 8 * time.Second,
					})

					return h
				},
				dataFile: "./testr/micro-config-write-sucess.dat",
			},
			argBody:   `{"iniSendTime": "12:00", "endSendTime": "12:00", "wakeup": 10, "buffer": 10}`,
			resStatus: http.StatusOK,
		},
		{
			name: "Save. StatusBadRequest",
			field: fields{
				hubf:     func() micro.Hub { return mocks.NewHub(t) },
				dataFile: "./testr/micro-config-write-sucess.dat",
			},
			argBody:   "{",
			resStatus: http.StatusBadRequest,
		},
		{
			name: "Save. StatusInternalServerError",
			field: fields{
				hubf:     func() micro.Hub { return mocks.NewHub(t) },
				dataFile: "./not_exist/micro-config-write.dat",
			},
			argBody:   `{"iniSendTime": "12:00", "endSendTime": "12:00", "wakeup": 10, "buffer": 10}`,
			resStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			body := strings.NewReader(tt.argBody)
			req := httptest.NewRequest(http.MethodPost, "/config", body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			cf := &web.ConfigWeb{
				Log:    zap,
				MicroR: &micro.ConfigRead{},
				MicroW: &micro.ConfigWrite{
					Log:      zap,
					MControl: &micro.Controller{},
					Hub:      tt.field.hubf(),
					Config:   config.Default(),
					DataFile: tt.field.dataFile,
				},
			}

			_ = cf.Save(c)

			if rec.Code == http.StatusOK {
				if err := os.Remove(cf.MicroW.DataFile); err != nil {
					assert.Error(t, err, "Removing micro config file created")
				}
			}

			assert.Equal(t, tt.resStatus, rec.Code)
		})
	}
}
