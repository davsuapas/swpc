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

package iot_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/iot"
	"go.uber.org/zap"
)

func TestAuth_Token(t *testing.T) {
	t.Parallel()

	type args struct {
		sID string
	}

	tests := []struct {
		name   string
		args   args
		status int
	}{
		{
			name: "Token. Success",
			args: args{
				sID: "",
			},
			status: http.StatusOK,
		},
		{
			name: "Token. Invalid",
			args: args{
				sID: "invalid",
			},
			status: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/token/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/token/:" + iot.ClientIDName)
			c.SetParamNames(iot.ClientIDName)
			c.SetParamValues(tt.args.sID)

			o := iot.NewAuth(zap.NewExample(), config.API{})

			_ = o.Token(c)

			assert.Equal(t, tt.status, rec.Code)

			if tt.status == http.StatusOK {
				assert.NotEmpty(t, rec.Body.String())
			}
		})
	}
}
