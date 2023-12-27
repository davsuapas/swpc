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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/internal/web/mocks"
	"github.com/swpoolcontroller/pkg/auth"
	"go.uber.org/zap"
)

var (
	errToken = errors.New("token error")
)

func TestAuthFlowDev_Login_Should_Return_StatusFound(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	o := &web.AuthFlowDev{
		Log:    zap.NewExample(),
		Config: config.Default(),
	}

	_ = o.Login(c)

	assert.Equal(t, http.StatusFound, rec.Code)

	r := rec.Result()
	defer r.Body.Close()
	assert.Equal(t, 2, len(r.Cookies()), "Cookies")
}

func TestAuthFlowDev_Logout_Should_Return_StatusFound(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	clientID := "1234"

	cookie := &http.Cookie{
		Name:  web.WSClientIDName,
		Value: clientID,
	}

	req.AddCookie(cookie)

	hub := mocks.NewHub(t)
	hub.On("UnregisterClient", clientID)

	o := &web.AuthFlowDev{
		Log:    zap.NewExample(),
		Hub:    hub,
		Config: config.Default(),
	}

	_ = o.Logout(c)

	assert.Equal(t, http.StatusFound, rec.Code)

	r := rec.Result()
	defer r.Body.Close()
	assert.Equal(t, 1, len(r.Cookies()), "Cookies")

	hub.AssertExpectations(t)
}

func TestAuthFlow_Login(t *testing.T) {
	t.Parallel()

	type args struct {
		state string
		code  string
	}

	type want struct {
		statusCode int
		cookies    int
		redirect   string
	}

	type mock struct {
		use   bool
		token *jwt.Token
		err   error
	}

	state, err := auth.EncodeState(
		[]byte(config.Default().SecretKey),
		[]byte("state"))
	if err != nil {
		t.Error("Encode state error")

		return
	}

	tests := []struct {
		name string
		args args
		mock mock
		want want
	}{
		{
			name: `Login with wrong state. 
						 It should return StatusFound with RedirectLoginURI`,
			args: args{
				state: "123",
				code:  "123",
			},
			mock: mock{
				use: false,
			},
			want: want{
				statusCode: http.StatusFound,
				cookies:    0,
				redirect:   web.RedirectErrorAuth,
			},
		},
		{
			name: `Login with wrong token. 
						 It should return StatusFound with RedirectLoginURI`,
			args: args{
				state: state,
				code:  "123",
			},
			mock: mock{
				use:   true,
				token: &jwt.Token{},
				err:   errToken,
			},
			want: want{
				statusCode: http.StatusFound,
				cookies:    0,
				redirect:   web.RedirectErrorAuth,
			},
		},
		{
			name: "Login. It should return StatusFound with RedirectLoginOk",
			args: args{
				state: state,
				code:  "123",
			},
			mock: mock{
				use:   true,
				token: &jwt.Token{Valid: true},
			},
			want: want{
				statusCode: http.StatusFound,
				cookies:    3,
				redirect:   web.RedirectLoginOk,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			queryParams := url.Values{}
			queryParams.Set("state", tt.args.state)
			queryParams.Set("code", tt.args.code)
			req := httptest.NewRequest(
				http.MethodGet,
				"/login?"+queryParams.Encode(),
				nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			oa := mocks.NewOAuth2(t)

			o := &web.AuthFlow{
				Log:     zap.NewExample(),
				Service: oa,
				Config:  config.Default(),
			}

			param := auth.OA2TokenInput{
				URL:         o.Config.Auth.TokenURL,
				Code:        tt.args.code,
				RedirectURI: o.Config.AuthRedirectURI(web.RedirectLogin),
			}
			if tt.mock.use {
				oa.On("Token", param).Return(tt.mock.token, tt.mock.err)
			}

			_ = o.Login(c)

			assert.Equal(t, tt.want.statusCode, rec.Code)

			r := rec.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.want.cookies, len(r.Cookies()), "Cookies")
			assert.Equal(t, tt.want.redirect, rec.Header().Get("Location"))

			oa.AssertExpectations(t)
		})
	}
}

func TestAuthFlow_Logout(t *testing.T) {
	t.Parallel()

	type args struct {
		clientName string
		clientID   string
	}

	type want struct {
		statusCode int
		cookies    int
	}

	tests := []struct {
		name       string
		args       args
		mockHubUse bool
		want       want
	}{
		{
			name: `Logout without WSClientIDName cookie.
			 		 	 it should return StatusInternalServerError`,
			args: args{
				clientName: "error",
			},
			mockHubUse: false,
			want: want{
				statusCode: http.StatusInternalServerError,
				cookies:    0,
			},
		},
		{
			name: "Logout. it should return StatusOK",
			args: args{
				clientName: web.WSClientIDName,
				clientID:   "123",
			},
			mockHubUse: true,
			want: want{
				statusCode: http.StatusOK,
				cookies:    2,
			},
		}}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/logout", nil)
			co := &http.Cookie{
				Name:  tt.args.clientName,
				Value: tt.args.clientID,
			}
			req.AddCookie(co)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := mocks.NewHub(t)

			o := &web.AuthFlow{
				Log:    zap.NewExample(),
				Hub:    h,
				Config: config.Default(),
			}

			if tt.mockHubUse {
				h.On("UnregisterClient", tt.args.clientID)
			}

			_ = o.Logout(c)

			assert.Equal(t, tt.want.statusCode, rec.Code)

			r := rec.Result()
			defer r.Body.Close()
			assert.Equal(t, tt.want.cookies, len(r.Cookies()), "Cookies")

			h.AssertExpectations(t)
		})
	}
}
