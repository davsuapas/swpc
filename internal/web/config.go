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

package web

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

// ConfigWeb manages the web configuration
type ConfigWeb struct {
	Log    *zap.Logger
	Microc *config.MicroConfigController
	Hub    *sockets.Hub
	Config config.Config
}

// Load loads the configuration saved into disk file
func (cf *ConfigWeb) Load(ctx echo.Context) error {
	data, err := cf.Microc.Read()
	if errors.Is(err, os.ErrNotExist) {
		return ctx.NoContent(http.StatusNotFound)
	}

	if err != nil {
		cf.Log.Error("Loading configuration", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, data)
}

// Save saves the configuration to disk
func (cf *ConfigWeb) Save(ctx echo.Context) error {
	var conf config.MicroConfig

	if err := ctx.Bind(&conf); err != nil {
		cf.Log.Error("Getting the configuration of the request body", zap.Error(err))

		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := cf.Microc.Save(conf); err != nil {
		cf.Log.Error("Saving config request", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	cf.Hub.Config(sockets.Config{
		CommLatency: time.Duration(cf.Config.CommLatencyTime),
		Buffer:      time.Duration(conf.Buffer),
	})

	return ctx.NoContent(http.StatusOK)
}
