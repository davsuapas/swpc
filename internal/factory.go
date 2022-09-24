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
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Factory is the objects factory of the app
type Factory struct {
	Config Config
	Webs   *echo.Echo
	Logger *zap.Logger
}

// NewFactory creates the horizontal services of the app
func NewFactory() *Factory {
	c := loadConfig()

	return &Factory{
		Config: c,
		Webs:   newWebServer(),
		Logger: newLogger(c),
	}
}

func newLogger(c Config) *zap.Logger {
	var log zap.Config

	if c.Development {
		log = zap.NewDevelopmentConfig()
	} else {
		log = zap.NewProductionConfig()
	}
	log.Level = zap.NewAtomicLevelAt(zapcore.Level(c.Level))
	log.Encoding = c.Encoding

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
