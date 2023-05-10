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
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/internal/web/mocks"
	"go.uber.org/zap"
)

func TestWS_Register_WS_Should_Return_StatusOk(t *testing.T) {
	t.Parallel()

	cstatusCode := make(chan int)

	hubm := mocks.NewHub(t)
	hubm.On("Register", mock.AnythingOfType("sockets.Client"))

	regh := func(w http.ResponseWriter, r *http.Request) {
		ws := web.NewWS(zap.NewExample(), config.Default().Web, hubm)
		c := echo.New().NewContext(r, w)
		_ = ws.Register(c)
		cstatusCode <- c.Response().Status
	}

	s := httptest.NewServer(http.HandlerFunc(regh))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	cookie := http.Cookie{Name: web.WSClientIDName, Value: "1234", Path: "/", HttpOnly: true, Secure: false}
	header := make(http.Header)
	header.Add("Cookie", cookie.String())

	ws, r, err := websocket.DefaultDialer.Dial(u, header)

	if err != nil {
		t.Fatalf("%v", err)
	}
	defer r.Body.Close()
	defer ws.Close()

	assert.Equal(t, http.StatusOK, <-cstatusCode)

	hubm.AssertExpectations(t)
}

func TestWS_Register_Http_Should_Return_StatusBadRequest(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ws := web.NewWS(zap.NewExample(), config.Default().Web, mocks.NewHub(t))

	_ = ws.Register(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWS_Register_WS_Without_Socket_Should_Return_StatusInternalServerError(t *testing.T) {
	t.Parallel()

	cstatusCode := make(chan int)

	regh := func(w http.ResponseWriter, r *http.Request) {
		ws := web.NewWS(zap.NewExample(), config.Default().Web, mocks.NewHub(t))
		c := echo.New().NewContext(r, w)
		_ = ws.Register(c)
		cstatusCode <- c.Response().Status
	}

	s := httptest.NewServer(http.HandlerFunc(regh))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, r, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer r.Body.Close()
	defer ws.Close()

	assert.Equal(t, http.StatusInternalServerError, <-cstatusCode)
}
