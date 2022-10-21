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
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WebHandler Web handler to router
type WebHandler struct {
	Login  *web.Login
	Config *web.ConfigWeb
}

// Factory is the objects factory of the app
type Factory struct {
	Config config.Config
	Webs   *echo.Echo
	Log    *zap.Logger

	WebHandler *WebHandler
}

// NewFactory creates the horizontal services of the app
func NewFactory() *Factory {
	cnf := config.LoadConfig()

	log := newLogger(cnf)

	return &Factory{
		Config: cnf,
		Webs:   newWebServer(),
		Log:    log,
		WebHandler: &WebHandler{
			Login: web.NewLogin(log),
			Config: &web.ConfigWeb{
				Log: log,
				Microc: &config.MicroConfig{
					Log:      log,
					DataPath: cnf.DataPath,
				},
				Cnf: cnf,
			},
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
