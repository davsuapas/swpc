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
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/micro"
	"go.uber.org/zap"
)

const (
	errloadConfig    = "Loading configuration"
	errGettingConfig = "Getting the configuration of the request body"
	errSavingConfig  = "Saving config request"
)

// ConfigWeb manages the web configurationhttps://github.com/aws/aws-sdk-go-v2/tree/main/service/dynamodb
type ConfigWeb struct {
	Log    *zap.Logger
	MicroR micro.ConfigRead
	MicroW micro.ConfigWrite
}

// Load loads the configuration from disk file
func (cf *ConfigWeb) Load(ctx echo.Context) error {
	data, err := cf.MicroR.Read()
	if err != nil {
		cf.Log.Error(errloadConfig, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, data)
}

// Save saves the configuration to disk
func (cf *ConfigWeb) Save(ctx echo.Context) error {
	var conf micro.Config

	if err := ctx.Bind(&conf); err != nil {
		cf.Log.Error(errGettingConfig, zap.Error(err))

		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := cf.MicroW.Save(conf); err != nil {
		cf.Log.Error(errSavingConfig, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
