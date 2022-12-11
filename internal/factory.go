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

package internal

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/api"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/sockets"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// APIHandler Micro API handler
type APIHandler struct {
	OAuth  *api.OAuth
	Stream *api.Stream
}

// WebHandler Web handler
type WebHandler struct {
	Login  *web.Login
	Config *web.ConfigWeb
	WS     *web.WS
}

// Factory is the objects factory of the app
type Factory struct {
	Config config.Config
	Webs   *echo.Echo
	Log    *zap.Logger

	Hubt *HubTrace
	Hub  *sockets.Hub

	WebHandler *WebHandler
	APIHandler *APIHandler
}

// NewFactory creates the horizontal services of the app
func NewFactory() *Factory {
	cnf := config.LoadConfig()

	log := newLogger(cnf)

	hubt := NewHubTrace(log)
	hub := sockets.NewHub(
		sockets.Config{
			CommLatency: time.Duration(cnf.CommLatencyTime) * time.Second,
			Buffer:      5 * time.Second,
		},
		hubt.Infos,
		hubt.Errors)

	mcontrol := &micro.Controller{
		Log:            log,
		Hub:            hub,
		Config:         config.MicroConfig{},
		CheckTransTime: uint8(cnf.CheckTransTime),
	}

	return &Factory{
		Config: cnf,
		Webs:   newWebServer(),
		Log:    log,
		Hubt:   hubt,
		Hub:    hub,
		WebHandler: &WebHandler{
			Login: web.NewLogin(log, cnf.WebConfig, hub),
			Config: &web.ConfigWeb{
				Log: log,
				Hub: hub,
				Microc: &config.MicroConfigController{
					Log:      log,
					DataPath: cnf.DataPath,
				},
				Config: cnf,
			},
			WS: web.NewWS(log, cnf.WebConfig, hub),
		},
		APIHandler: &APIHandler{
			OAuth:  api.NewOAuth(log, cnf.APIConfig),
			Stream: api.NewStream(mcontrol),
		},
	}
}

func newLogger(ctx config.Config) *zap.Logger {
	var log zap.Config

	if ctx.Development {
		log = zap.NewDevelopmentConfig()
	} else {
		log = zap.NewProductionConfig()
	}

	log.Level = zap.NewAtomicLevelAt(zapcore.Level(ctx.Level))
	log.Encoding = ctx.Encoding

	l, err := log.Build()
	if err != nil {
		panic(strings.Concat("Error creating zap logger. Description: ", err.Error()))
	}

	return l
}

func newWebServer() *echo.Echo {
	e := echo.New()

	return e
}
