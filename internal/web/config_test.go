/*
 *   Copyright (c) 2022 ELIPCERO
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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swpoolcontroller/internal/config"
	iotc "github.com/swpoolcontroller/internal/iot"
	"github.com/swpoolcontroller/internal/iot/mocks"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/iot"
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
				body: "{\"iniSendTime\":\"10:00\",\"endSendTime\":\"21:01\"," +
					"\"wakeup\":16,\"buffer\":3," +
					"\"calibrationOrp\":0,\"calibrationPh\":0," +
					"\"calibratingOrp\":false,\"targetOrp\":0," +
					"\"calibratingPh\":false,\"targetPh\":0," +
					"\"stabilizationTime\":0}\n",
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/config", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			cf := &web.ConfigWeb{
				Log: zap,
				MicroR: &iotc.FileConfigRead{
					Log:      zap,
					DataFile: tt.dataFile,
				},
				MicroW: &iotc.FileConfigWrite{},
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
		hubf     func() iotc.Hub
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
				hubf: func() iotc.Hub {
					h := mocks.NewHub(t)
					h.On("Config", iot.DeviceConfig{
						WakeUpTime:         10,
						CollectMetricsTime: 1000,
						Buffer:             10,
						IniSendTime:        "12:00",
						EndSendTime:        "12:00",
						CalibrationORP:     1132.12,
						CalibrationPH:      1.12,
						CalibratingORP:     true,
						TargetORP:          450.10,
						CalibratingPH:      true,
						TargetPH:           7.2,
						StabilizationTime:  20,
					})

					return h
				},
				dataFile: "./testr/micro-config-write-sucess.dat",
			},
			argBody: `{"iniSendTime":"12:00","endSendTime":"12:00",
			 "wakeup":10,"buffer":10,
			 "calibrationOrp":1132.12,"calibrationPh":1.12,
			 "calibratingOrp":true,"targetOrp":450.10,
			 "calibratingPh":true,"targetPh":7.2,
			 "stabilizationTime":20}`,
			resStatus: http.StatusOK,
		},
		{
			name: "Save. StatusBadRequest",
			field: fields{
				hubf:     func() iotc.Hub { return mocks.NewHub(t) },
				dataFile: "./testr/micro-config-write-sucess.dat",
			},
			argBody:   "{",
			resStatus: http.StatusBadRequest,
		},
		{
			name: "Save. StatusInternalServerError",
			field: fields{
				hubf:     func() iotc.Hub { return mocks.NewHub(t) },
				dataFile: "./not_exist/micro-config-write.dat",
			},
			argBody: `{"iniSendTime": "12:00", "endSendTime": "12:00",
			 "wakeup": 10, "buffer": 10}`,
			resStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
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
				MicroR: &iotc.FileConfigRead{},
				MicroW: &iotc.FileConfigWrite{
					Log:      zap,
					Hub:      tt.field.hubf(),
					Config:   config.Default(),
					DataFile: tt.field.dataFile,
				},
			}

			_ = cf.Save(c)

			if rec.Code == http.StatusOK {
				if err := os.Remove(tt.field.dataFile); err != nil {
					require.Error(t, err, "Removing micro config file created")
				}
			}

			assert.Equal(t, tt.resStatus, rec.Code)
		})
	}
}
