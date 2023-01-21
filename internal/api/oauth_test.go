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

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/api"
	"github.com/swpoolcontroller/internal/config"
	"go.uber.org/zap"
)

func TestOAuth_Token(t *testing.T) {
	t.Parallel()

	type args struct {
		si string
	}

	tests := []struct {
		name   string
		args   args
		status int
	}{
		{
			name: "Token. Success",
			args: args{
				si: "sw3kf$fekdy56dfh",
			},
			status: http.StatusOK,
		},
		{
			name: "Token. Invalid",
			args: args{
				si: "tinvalid",
			},
			status: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/token/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/token/:secretID")
			c.SetParamNames("secretID")
			c.SetParamValues(tt.args.si)

			o := api.NewOAuth(zap.NewExample(), config.APIConfig{SessionExpiration: 1})

			_ = o.Token(c)

			assert.Equal(t, tt.status, rec.Code)
			if tt.status == http.StatusOK {
				assert.NotEmpty(t, rec.Body.String())
			}
		})
	}
}
