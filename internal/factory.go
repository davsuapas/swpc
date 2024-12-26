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

package internal

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsc "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/ai"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/hub"
	iotc "github.com/swpoolcontroller/internal/iot"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/auth"
	"github.com/swpoolcontroller/pkg/crypto"
	"github.com/swpoolcontroller/pkg/iot"
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
	inflog          = "Log configuration"
)

// APIHandler is the device API handler
type APIHandler struct {
	Auth *iotc.Auth
	WS   *iotc.WS
}

// WebHandler Web handler
type WebHandler struct {
	AppConfig  web.AppConfigurator
	Auth       web.Auth
	Config     *web.ConfigWeb
	Sample     *web.SampleWeb
	Prediction *web.PredictionWeb
	WS         *web.WS
}

// Factory is the objects factory of the app
type Factory struct {
	Config config.Config

	Webs *echo.Echo
	Log  *zap.Logger

	JWT *auth.JWT

	Hubt *hub.Trace
	Hub  *iot.Hub

	WebHandler *WebHandler
	APIHandler *APIHandler
}

// NewFactory creates the horizontal services of the app
func NewFactory() *Factory {
	cnf := config.LoadConfig()
	awscnf := newAWSConfig(cnf)

	log := newLogger(cnf)

	if !cnf.Hide {
		log.Info(infConfigLoaded, zap.String("Config", cnf.String()))
	}

	s := secretProvider(cnf, awscnf)
	config.ApplySecret(s, &cnf)

	mconfigRead := microConfigRead(cnf, awscnf, log)

	microc, err := mconfigRead.Read()
	if err != nil {
		log.Panic(errReadConfig)
	}

	loc, err := time.LoadLocation(cnf.Location.Zone)
	if err != nil {
		log.Panic(err.Error())
	}

	hubt, hub := newHub(log, cnf, microc, loc)

	jwt := &auth.JWT{
		JWKFetch: auth.NewJWKFetch(cnf.Auth.JWKURL),
	}

	mconfigWrite := microConfigWrite(cnf, awscnf, log, hub)

	return &Factory{
		Config:     cnf,
		Webs:       echo.New(),
		Log:        log,
		JWT:        jwt,
		Hubt:       hubt,
		Hub:        hub,
		WebHandler: newWeb(log, cnf, awscnf, hub, jwt, mconfigRead, mconfigWrite),
		APIHandler: &APIHandler{
			Auth: iotc.NewAuth(log, cnf.API),
			WS:   iotc.NewWS(log, hub),
		},
	}
}

func microConfigRead(
	cnf config.Config,
	cnfaws *awsConfig,
	log *zap.Logger) iotc.ConfigRead {
	//
	if cnf.Data.Provider == config.CloudDataProvider &&
		cnf.Cloud.Provider != config.NoneCloudProvider &&
		cnf.Data.AWS.ConfigTableName != "" {
		return iotc.NewAWSConfigRead(
			log,
			cnfaws.get(),
			cnf.Data.AWS.ConfigTableName)
	}

	if cnf.Data.Provider == config.FileDataProvider &&
		cnf.Data.File.ConfigFile != "" {
		return &iotc.FileConfigRead{
			Log:      log,
			DataFile: cnf.Data.File.ConfigFile,
		}
	}

	return &iotc.DefaultConfigRead{
		Log: log,
	}
}

func microConfigWrite(
	cnf config.Config,
	cnfaws *awsConfig,
	log *zap.Logger,
	hub *iot.Hub) iotc.ConfigWrite {
	//
	if cnf.Data.Provider == config.CloudDataProvider &&
		cnf.Cloud.Provider != config.NoneCloudProvider &&
		cnf.Data.AWS.ConfigTableName != "" {
		return iotc.NewAWSConfigWrite(
			log,
			hub,
			cnf,
			cnfaws.get(), cnf.Data.AWS.ConfigTableName)
	}

	if cnf.Data.Provider == config.FileDataProvider &&
		cnf.Data.File.ConfigFile != "" {
		return &iotc.FileConfigWrite{
			Log:      log,
			Hub:      hub,
			Config:   cnf,
			DataFile: cnf.Data.File.ConfigFile,
		}
	}

	return &iotc.DefaultConfigSave{
		Log: log,
	}
}

func secretProvider(cnf config.Config, cnfaws *awsConfig) config.Secret {
	if cnf.Cloud.Provider == config.CloudAWSProvider && len(cnf.Secret.Name) > 0 {
		return crypto.NewAWSSecret(cnfaws.get())
	}

	return &config.DummySecret{}
}

//nolint:funlen
func newWeb(
	log *zap.Logger,
	cnf config.Config,
	cnfaws *awsConfig,
	hub *iot.Hub,
	jwt *auth.JWT,
	mconfigRead iotc.ConfigRead,
	mconfigWrite iotc.ConfigWrite) *WebHandler {
	//
	var oauth2 web.Auth

	var appConfig web.AppConfigurator

	repoSample := buildSampleRepo(cnf, cnfaws, log)

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
			Log:    log,
			Hub:    hub,
			Config: cnf,
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
		Sample: &web.SampleWeb{
			Log:  log,
			Repo: repoSample,
		},
		Prediction: &web.PredictionWeb{
			Preder: &ai.Prediction{
				Log: log,
			},
			Log: log,
		},
		WS: web.NewWS(log, cnf.Web, hub),
	}
}

func buildSampleRepo(
	cnf config.Config,
	cnfaws *awsConfig,
	log *zap.Logger) ai.SampleRepo {
	//
	switch cnf.Data.Provider {
	case config.CloudDataProvider:
		if cnf.Data.AWS.SamplesTableName != "" {
			return ai.NewSampleAWSDynamoRepo(
				cnfaws.get(),
				log,
				cnf.Data.AWS.SamplesTableName)
		}
	case config.FileDataProvider:
		if cnf.Data.File.SampleFile != "" {
			return &ai.SampleFileRepo{
				Log:      log,
				FileName: cnf.Data.File.SampleFile,
			}
		}
	case config.NoneDataProvider:
		return &web.SampleDummyRepo{Log: log}
	}

	return &web.SampleDummyRepo{Log: log}
}

func newHub(
	log *zap.Logger,
	config config.Config,
	microc iotc.Config,
	loc *time.Location) (*hub.Trace, *iot.Hub) {
	//
	hubTraceLevel := iot.NoneLevel

	switch log.Level() { //nolint:exhaustive
	case zap.DebugLevel:
		hubTraceLevel = iot.DebugLevel
	case zap.InfoLevel:
		hubTraceLevel = iot.InfoLevel
	case zap.WarnLevel:
		hubTraceLevel = iot.WarnLevel
	}

	hubt := hub.NewTrace(log)
	hub := iot.NewHub(
		iot.Config{
			DeviceConfig: iot.DeviceConfig{
				WakeUpTime:         microc.Wakeup,
				CollectMetricsTime: config.CollectMetricsTime,
				Buffer:             microc.Buffer,
				IniSendTime:        microc.IniSendTime,
				EndSendTime:        microc.EndSendTime,
			},
			Location:         loc,
			CommLatency:      time.Duration(config.CommLatencyTime) * time.Second,
			TaskTime:         time.Duration(config.TaskTime) * time.Second,
			NotificationTime: time.Duration(config.NotificationTime) * time.Second,
			HeartbeatConfig: iot.HeartbeatConfig{
				HeartbeatInterval: time.Duration(config.HeartbeatInterval) *
					time.Second,
				HeartbeatPingTime: time.Duration(config.HeartbeatPingTime) *
					time.Second,
				HeartbeatTimeoutCount: config.HeartbeatTimeoutCount,
			},
		},
		hubTraceLevel,
		hubt.Trace,
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

	// Set the time format to ISO8601 for better readability
	log.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := log.Build()
	if err != nil {
		panic(
			strings.Format(
				errCreateZap,
				strings.FMTValue("Description",
					err.Error())))
	}

	l.Info(
		inflog,
		zap.Bool("Develop", ctx.Development),
		zap.Int8("Level", ctx.Level),
		zap.String("Encoding", log.Encoding),
	)

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
