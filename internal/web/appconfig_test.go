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
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/web"
	"go.uber.org/zap"
)

func TestAppConfig_Load(t *testing.T) {
	t.Parallel()

	type res struct {
		status int
		body   string
	}

	tests := []struct {
		name      string
		argConfig config.Config
		res       res
	}{
		{
			name: "Load config. Error encrypt key length",
			argConfig: config.Config{
				Web: config.Web{},
			},
			res: res{
				status: http.StatusInternalServerError,
				body:   "",
			},
		},
		{
			name:      "Load config. Status Code = 200",
			argConfig: config.Default(),
			res: res{
				status: http.StatusOK,
				body: "{\"authLoginUrl\":\"\",\"authLogoutUrl\":\"\"," +
					"\"checkAuthName\":\"IsAuth\"}\n",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/appconfig", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ac := web.AppConfig{
				Log:    zap.NewExample(),
				Config: tt.argConfig,
			}

			_ = ac.Load(ctx)

			assert.Equal(t, tt.res.status, rec.Code)
			assert.Equal(t, tt.res.body, rec.Body.String())
		})
	}
}

func TestAppConfigDev_Load(t *testing.T) {
	t.Parallel()

	type res struct {
		status int
		body   string
	}

	tests := []struct {
		name      string
		argConfig config.Config
		res       res
	}{
		{
			name:      "Load config. Status Code = 200",
			argConfig: config.Default(),
			res: res{
				status: http.StatusOK,
				body: "{\"authLoginUrl\":\"/auth/login\"," +
					"\"authLogoutUrl\":\"/auth/logout\",\"checkAuthName\":\"IsAuth\"}\n",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/appconfig", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ac := web.AppConfigDev{
				Log:    zap.NewExample(),
				Config: tt.argConfig,
			}

			_ = ac.Load(ctx)

			assert.Equal(t, tt.res.status, rec.Code)
			assert.Equal(t, tt.res.body, rec.Body.String())
		})
	}
}
