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
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/api"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/hub"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/auth"
	"github.com/swpoolcontroller/pkg/sockets"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	errReadConfig = "Reading the configuration of the micro controller from config file"
	errCreateZap  = "Error creating zap logger"
)

const dataFile = "micro-config.dat"

// APIHandler Micro API handler
type APIHandler struct {
	Auth   *api.Auth
	Stream *api.Stream
}

// WebHandler Web handler
type WebHandler struct {
	AppConfig web.AppConfigurator
	Auth      web.Auth
	Config    *web.ConfigWeb
	WS        *web.WS
}

// Factory is the objects factory of the app
type Factory struct {
	Config config.Config

	Webs *echo.Echo
	Log  *zap.Logger

	JWT *auth.JWT

	Hubt *hub.Trace
	Hub  *sockets.Hub

	WebHandler *WebHandler
	APIHandler *APIHandler
}

// NewFactory creates the horizontal services of the app
func NewFactory() *Factory {
	config := config.LoadConfig()

	log := newLogger(config)

	dataFile := path.Join(config.DataPath, dataFile)

	mconfigRead := &micro.ConfigRead{
		Log:      log,
		DataFile: dataFile,
	}

	configm, err := mconfigRead.Read()
	if err != nil {
		log.Panic(errReadConfig)
	}

	hubt, hub := newHub(log, config, configm)

	mcontrol := &micro.Controller{
		Log:                log,
		Hub:                hub,
		Config:             configm,
		CheckTransTime:     uint8(config.CheckTransTime),
		CollectMetricsTime: uint16(config.CollectMetricsTime),
	}

	mconfigWrite := &micro.ConfigWrite{
		Log:      log,
		MControl: mcontrol,
		Hub:      hub,
		Config:   config,
		DataFile: dataFile,
	}

	jwt := &auth.JWT{
		JWKFetch: auth.NewJWKFetch(config.Auth.JWKURL),
	}

	return &Factory{
		Config:     config,
		Webs:       echo.New(),
		Log:        log,
		JWT:        jwt,
		Hubt:       hubt,
		Hub:        hub,
		WebHandler: newWeb(log, config, hub, jwt, mconfigRead, mconfigWrite),
		APIHandler: &APIHandler{
			Auth:   api.NewAuth(log, config.API),
			Stream: api.NewStream(mcontrol),
		},
	}
}

func newWeb(
	log *zap.Logger,
	cnf config.Config,
	hub *sockets.Hub,
	jwt *auth.JWT,
	mconfigRead *micro.ConfigRead,
	mconfigWrite *micro.ConfigWrite) *WebHandler {
	//
	var oauth2 web.Auth

	var appConfig web.AppConfigurator

	if cnf.Auth.Provider == config.AuthProviderOauth2 {
		oauth2 = &web.AuthFlow{
			Log: log,
			Service: &auth.OAuth2{
				ClientID: cnf.Auth.ClientID,
				JWT:      jwt,
			},
			Hub:    hub,
			Config: cnf,
		}

		appConfig = &web.AppConfig{
			Log:    log,
			Config: cnf,
		}
	} else {
		log.Warn("Authentication has been configured in development mode. Never use this configuration in production.")

		oauth2 = &web.AuthFlowDev{
			Log:  log,
			Hub:  hub,
			Webc: cnf.Web,
		}

		appConfig = &web.AppConfigDev{
			Log:    log,
			Config: cnf,
		}
	}

	return &WebHandler{
		AppConfig: appConfig,
		Auth:      oauth2,
		Config: &web.ConfigWeb{
			Log:    log,
			MicroR: mconfigRead,
			MicroW: mconfigWrite,
		},
		WS: web.NewWS(log, cnf.Web, hub),
	}
}

func newHub(log *zap.Logger, config config.Config, configm micro.Config) (*hub.Trace, *sockets.Hub) {
	hubt := hub.NewTrace(log)
	hub := sockets.NewHub(
		sockets.Config{
			CommLatency:      time.Duration(config.CommLatencyTime) * time.Second,
			Buffer:           time.Duration(configm.Buffer) * time.Second,
			TaskTime:         time.Duration(config.TaskTime) * time.Second,
			NotificationTime: time.Duration(config.NotificationTime) * time.Second,
		},
		hubt.Info,
		hubt.Error)

	return hubt, hub
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
		panic(strings.Format(errCreateZap, strings.FMTValue("Description", err.Error())))
	}

	return l
}
