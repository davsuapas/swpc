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
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/internal/web/mocks"
	"go.uber.org/zap"
)

func TestLogin_Logoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldHubf func() *mocks.Hub
		argCookie string
		resStatus int
	}{
		{
			name: "Logoff. StatusOk",
			fieldHubf: func() *mocks.Hub {
				h := mocks.NewHub(t)
				h.On("Unregister", mock.AnythingOfType("string"))

				return h
			},
			argCookie: "123",
			resStatus: http.StatusOK,
		},
		{
			name:      "Logoff. StatusInternalServerError",
			fieldHubf: func() *mocks.Hub { return mocks.NewHub(t) },
			argCookie: "",
			resStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/logoff", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if len(tt.argCookie) != 0 {
				cookie := &http.Cookie{
					Name:  web.TokenName,
					Value: tt.argCookie,
				}
				req.AddCookie(cookie)
			}

			mh := tt.fieldHubf()

			l := web.NewLogin(zap.NewExample(), config.WebConfig{}, mh)

			_ = l.Logoff(c)

			assert.Equal(t, tt.resStatus, rec.Code)

			mh.AssertExpectations(t)
		})
	}
}

func TestLogin_Submit(t *testing.T) {
	t.Parallel()

	type args struct {
		user string
		pass string
	}

	tests := []struct {
		name      string
		args      args
		resStatus int
	}{
		{
			name: "Login. StatusOk",
			args: args{
				user: "test",
				pass: "test",
			},
			resStatus: http.StatusOK,
		},
		{
			name: "Login. User not found returns StatusUnauthorized",
			args: args{
				user: "pepe",
				pass: "",
			},
			resStatus: http.StatusUnauthorized,
		},
		{
			name: "Login. Pass incorrect returns StatusUnauthorized",
			args: args{
				user: "test",
				pass: "pass",
			},
			resStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Form = url.Values{}
			req.Form.Add("email", tt.args.user)
			if len(tt.args.pass) != 0 {
				req.Form.Add("password", tt.args.pass)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			l := web.NewLogin(zap.NewExample(), config.WebConfig{}, nil)

			_ = l.Submit(c)

			assert.Equal(t, tt.resStatus, rec.Code)
		})
	}
}
