/*
 *   Copyright (c) 2022 ELIPCERO
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
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/auth"
	"go.uber.org/zap"
)

const (
	errEncodeState   = "Encode auth state"
	errMarshalConfig = "Marshal json config"
)

const (
	infLoadConfig = "Load app config"
)

type configDTO struct {
	AuthLoginURL  string `json:"authLoginUrl"`
	AuthLogoutURL string `json:"authLogoutUrl"`
	CheckAuthName string `json:"checkAuthName"`
	IOTConfig     bool   `json:"iotConfig"`
	AISample      bool   `json:"aiSample"`
}

type AppConfigurator interface {
	Load(ctx echo.Context) error
}

// AppConfig loads the app config to the UI. Therefore, the app config
// is centred in the server config
type AppConfig struct {
	Log    *zap.Logger
	Config config.Config
}

// Load loads the app configuration
func (c *AppConfig) Load(ctx echo.Context) error {
	state := xid.New().String()

	statec, err := auth.EncodeState([]byte(c.Config.Web.SecretKey), []byte(state))
	if err != nil {
		c.Log.Error(errEncodeState, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	config := configDTO{
		AuthLoginURL: auth.Oauth2URL(
			c.Config.Auth.LoginURL,
			c.Config.Auth.ClientID,
			c.Config.AuthRedirectURI(RedirectLogin),
			statec),
		AuthLogoutURL: auth.Oauth2URL(
			c.Config.Auth.LogoutURL,
			c.Config.Auth.ClientID,
			c.Config.AuthRedirectURI(RedirectLogout),
			""),
		CheckAuthName: AuthCheckName,
		IOTConfig:     c.Config.IOT.ConfigUI,
		AISample:      c.Config.IOT.SampleUI,
	}

	sconfig, err := json.Marshal(config)
	if err != nil {
		c.Log.Error(errMarshalConfig, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	c.Log.Info(infLoadConfig, zap.String("AppConfig", string(sconfig)))

	return ctx.JSON(http.StatusOK, config)
}

// AppConfig loads the app config to the UI for develepment mode
type AppConfigDev struct {
	Log    *zap.Logger
	Config config.Config
}

// Load loads the app configuration setting up authentication
// in development mode
func (c *AppConfigDev) Load(ctx echo.Context) error {
	config := configDTO{
		AuthLoginURL:  RedirectLogin,
		AuthLogoutURL: RedirectLogout,
		CheckAuthName: AuthCheckName,
		IOTConfig:     c.Config.IOT.ConfigUI,
		AISample:      c.Config.IOT.SampleUI,
	}

	sconfig, err := json.Marshal(config)
	if err != nil {
		c.Log.Error(errMarshalConfig, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	c.Log.Info(infLoadConfig, zap.String("AppConfig", string(sconfig)))

	return ctx.JSON(http.StatusOK, config)
}
