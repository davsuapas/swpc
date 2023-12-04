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
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsc "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/api"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/hub"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/auth"
	"github.com/swpoolcontroller/pkg/crypto"
	"github.com/swpoolcontroller/pkg/sockets"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	errReadConfig = "Reading the configuration of the micro controller " +
		"from config file"
	errCreateZap = "Error creating zap logger"
	errAWSConfig = "Error creating secret maanger"
)

const (
	infConfigLoaded = "The configuration is loaded"
)

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
	cnf := config.LoadConfig()
	awscnf := newAWSConfig(cnf)

	log := newLogger(cnf)

	log.Info(infConfigLoaded, zap.String("Config", cnf.String()))

	s := secretProvider(cnf, awscnf)
	config.ApplySecret(log, s, &cnf)

	mconfigRead := microConfigRead(cnf, awscnf, log)

	microc, err := mconfigRead.Read()
	if err != nil {
		log.Panic(errReadConfig)
	}

	hubt, hub := newHub(log, cnf, microc)

	loc, err := time.LoadLocation(cnf.Location.Zone)
	if err != nil {
		log.Panic(err.Error())
	}

	mcontrol := &micro.Controller{
		Location:           loc,
		Log:                log,
		Hub:                hub,
		Config:             microc,
		CheckTransTime:     uint8(cnf.CheckTransTime),
		CollectMetricsTime: uint16(cnf.CollectMetricsTime),
	}

	jwt := &auth.JWT{
		JWKFetch: auth.NewJWKFetch(cnf.Auth.JWKURL),
	}

	mconfigWrite := microConfigWrite(cnf, awscnf, log, mcontrol, hub)

	return &Factory{
		Config:     cnf,
		Webs:       echo.New(),
		Log:        log,
		JWT:        jwt,
		Hubt:       hubt,
		Hub:        hub,
		WebHandler: newWeb(log, cnf, hub, jwt, mconfigRead, mconfigWrite),
		APIHandler: &APIHandler{
			Auth:   api.NewAuth(log, cnf.API),
			Stream: api.NewStream(mcontrol),
		},
	}
}

func microConfigRead(
	cnf config.Config,
	cnfaws *awsConfig,
	log *zap.Logger) micro.ConfigRead {
	//
	if cnf.Data.Provider == config.CloudDataProvider &&
		cnf.Cloud.Provider != config.CloudNoProvider {
		return micro.NewAWSConfigRead(log, cnfaws.get(), cnf.Data.AWS.TableName)
	}

	return &micro.FileConfigRead{
		Log:      log,
		DataFile: cnf.Data.File.FilePath,
	}
}

func microConfigWrite(
	cnf config.Config,
	cnfaws *awsConfig,
	log *zap.Logger,
	mc *micro.Controller,
	hub *sockets.Hub) micro.ConfigWrite {
	//
	if cnf.Data.Provider == config.CloudDataProvider &&
		cnf.Cloud.Provider != config.CloudNoProvider {
		return micro.NewAWSConfigWrite(
			log,
			mc,
			hub,
			cnf,
			cnfaws.get(), cnf.Data.AWS.TableName)
	}

	return &micro.FileConfigWrite{
		Log:      log,
		MControl: mc,
		Hub:      hub,
		Config:   cnf,
		DataFile: cnf.Data.File.FilePath,
	}
}

func secretProvider(cnf config.Config, cnfaws *awsConfig) config.Secret {
	if cnf.Cloud.Provider == config.CloudAWSProvider && len(cnf.Secret.Name) > 0 {
		return crypto.NewAWSSecret(cnfaws.get())
	}

	return &config.DummySecret{}
}

func newWeb(
	log *zap.Logger,
	cnf config.Config,
	hub *sockets.Hub,
	jwt *auth.JWT,
	mconfigRead micro.ConfigRead,
	mconfigWrite micro.ConfigWrite) *WebHandler {
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
		log.Warn("Authentication has been configured in development mode. " +
			"Never use this configuration in production.")

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

func newHub(
	log *zap.Logger,
	config config.Config,
	configm micro.Config) (*hub.Trace, *sockets.Hub) {
	//
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
		panic(
			strings.Format(
				errCreateZap,
				strings.FMTValue("Description",
					err.Error())))
	}

	return l
}

type awsConfig struct {
	Config config.Config
	c      aws.Config
	load   bool
}

func newAWSConfig(cnf config.Config) *awsConfig {
	return &awsConfig{
		Config: cnf,
	}
}

func (ac *awsConfig) get() aws.Config {
	if ac.load {
		return ac.c
	}

	options := func(cfg *awsc.LoadOptions) error {
		if len(ac.Config.Cloud.AWS.Region) > 0 {
			cfg.Region = ac.Config.Cloud.AWS.Region
		}

		if len(ac.Config.Cloud.AWS.AKID) > 0 &&
			len(ac.Config.Cloud.AWS.SecretKey) > 0 {
			//
			cfg.Credentials = credentials.NewStaticCredentialsProvider(
				ac.Config.Cloud.AWS.AKID, ac.Config.Cloud.AWS.SecretKey, "")
		}

		return nil
	}

	cfg, err := awsc.LoadDefaultConfig(context.TODO(), options)

	if err != nil {
		panic(
			strings.Format(
				errAWSConfig,
				strings.FMTValue("Description", err.Error())))
	}

	ac.c = cfg
	ac.load = true

	return ac.c
}
