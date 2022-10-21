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
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"go.uber.org/zap"
)

// Config manages the web configuration
type ConfigWeb struct {
	Log    *zap.Logger
	Microc *config.MicroConfig
	Cnf    config.Config
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

	return ctx.JSONBlob(http.StatusOK, data)
}

// Save saves the configuration to disk
func (cf *ConfigWeb) Save(ctx echo.Context) error {
	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		cf.Log.Error("Getting the configuration of the request body", zap.Error(err))

		return ctx.NoContent(http.StatusBadRequest)
	}

	err = cf.Microc.Save(data)
	if err != nil {
		cf.Log.Error("Saving request", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
