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

package iot

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/iot"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const (
	errReadConfig = "Reading the configuration of the micro controller " +
		"from config file"
	errUnmarsConfig = "Unmarshalling the configuration " +
		"of the micro controller from config file"
	errMarshallConfig = "Marshalling the configuration of the micro controller: "
	errSaveConfig     = "Saving the configuration for the micro controller: "
)

const (
	infLoadingConfig     = "Loading configuration file"
	infConfigLoaded      = "Configuration file loaded"
	infSavingConfig      = "Saving configuration file"
	infdefaultReadConfig = "Reading the default configuration"
	infdefaultSaveConfig = "Save is not implemented"
	infConfig            = "Config"
	infFile              = "file"
)

const (
	dynamoDBTableKeyValue   = "1"
	dynamoDBTableKeyName    = "id"
	dynamoDBTableConfigName = "config"
)

type Config struct {
	// IniSendTime is the range for initiating metric sends
	IniSendTime string `json:"iniSendTime"`
	// EndSendTime is the range for ending metric sends
	EndSendTime string `json:"endSendTime"`
	// Wakeup is how often in minutes the micro-controller wakes up
	// to check for sending
	Wakeup uint8 `json:"wakeup"`
	// Buffer is the buffer in seconds to store metrics int the micro-controller.
	// It must be taken into account that if the buffer is for example 3 seconds,
	// double the buffer is stored, to avoid unnecessary waits in the web client.
	Buffer uint8 `json:"buffer"`
	// CalibratingORP is the flag to calibrate the ORP
	CalibratingORP bool `json:"calibratingOrp"`
	// TargetORP is the target value for the calibrating ORP
	TargetORP float32 `json:"targetOrp"`
	// CalibrationORP is the value for the calibration ORP
	// When set to calibrate with the flag it will be
	// the initial value of the calibration.
	// When not set to calibration mode, it will be the calibrated value
	// obtained from the calibration.
	CalibrationORP float32 `json:"calibrationOrp"`
	// StabilizationTimeORP is the time in seconds to stabilize
	// the calibration value
	StabilizationTimeORP int8 `json:"stabilizationTimeOrp"`
}

func DefaultConfig() Config {
	return Config{
		IniSendTime:          "09:00",
		EndSendTime:          "22:00",
		Wakeup:               30,
		Buffer:               3,
		CalibratingORP:       false,
		TargetORP:            469,
		CalibrationORP:       -320,
		StabilizationTimeORP: 20,
	}
}

// String returns struct as string
func (c *Config) String() string {
	r, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(r)
}

// ConfigRead reads the micro controller configuration
type ConfigRead interface {
	// Read reads the configuration
	Read() (Config, error)
}

// FileConfigWrite writes the micro controller configuration
type ConfigWrite interface {
	// Save saves the configuration
	Save(data Config) error
}

// DefaultConfigRead reads the default micro controller configuration
type DefaultConfigRead struct {
	Log *zap.Logger
}

// Read reads the default configuration
func (f *DefaultConfigRead) Read() (Config, error) {
	c := DefaultConfig()
	f.Log.Info(infdefaultReadConfig, zap.String(infConfig, c.String()))

	return DefaultConfig(), nil
}

// DefaultConfigSave not implement any action
type DefaultConfigSave struct {
	Log *zap.Logger
}

// Save not implement any action
func (f *DefaultConfigSave) Save(data Config) error {
	f.Log.Info(infdefaultSaveConfig, zap.String(infConfig, data.String()))

	return nil
}

// FileConfigRead reads the micro controller configuration from file
type FileConfigRead struct {
	Log      *zap.Logger
	DataFile string
}

// Read reads the configuration from disk.
// If the file not exists returns config default
func (c *FileConfigRead) Read() (Config, error) {
	c.Log.Info(infLoadingConfig, zap.String(infFile, c.DataFile))

	data, err := os.ReadFile(c.DataFile)
	if err != nil {
		c.Log.Error(errReadConfig, zap.String(infFile, c.DataFile), zap.Error(err))

		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}

		return Config{}, errors.Wrap(err, errReadConfig)
	}

	var mc Config

	if err := json.Unmarshal(data, &mc); err != nil {
		c.Log.Error(
			errUnmarsConfig,
			zap.String(infFile, c.DataFile),
			zap.Error(err))

		return Config{}, errors.Wrap(err, errUnmarsConfig)
	}

	c.Log.Info(infConfigLoaded, zap.String(infConfig, mc.String()))

	return mc, nil
}

// FileConfigWrite writes the micro controller configuration to file
type FileConfigWrite struct {
	Log      *zap.Logger
	Hub      Hub
	Config   config.Config
	DataFile string
}

// Save saves the configuration to disk
func (c FileConfigWrite) Save(data Config) error {
	c.Log.Info(
		infSavingConfig,
		zap.String(infConfig, data.String()), zap.String(infFile, c.DataFile))

	conf, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, strings.Concat(errMarshallConfig, c.DataFile))
	}

	if err := os.WriteFile(c.DataFile, conf, os.FileMode(0664)); err != nil {
		return errors.Wrap(err, strings.Concat(errSaveConfig, c.DataFile))
	}

	notifyHub(c.Config, data, c.Hub)

	return nil
}

// ConfigRead reads the micro controller configuration from AWS dynamodb
type AWSConfigRead struct {
	log       *zap.Logger
	tableName string
	client    *dynamodb.Client
}

// NewAWSConfigRead creates the micro controller configuration from AWS dynamodb
func NewAWSConfigRead(
	log *zap.Logger,
	cfg aws.Config,
	tableName string) *AWSConfigRead {
	//
	return &AWSConfigRead{
		log:       log,
		tableName: tableName,
		client:    dynamodb.NewFromConfig(cfg),
	}
}

// Read reads the configuration from dynamodb table.
// If the file not exists returns config default
func (c *AWSConfigRead) Read() (Config, error) {
	c.log.Info(infLoadingConfig, zap.String(infFile, c.tableName))

	res, err := c.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: dynamoDBTableKeyValue},
		},
	})
	if err != nil {
		c.log.Error(errReadConfig, zap.String(infFile, c.tableName), zap.Error(err))

		return Config{}, errors.Wrap(err, errReadConfig)
	}

	resMap := make(map[string]string)
	if err := attributevalue.UnmarshalMap(res.Item, &resMap); err != nil {
		c.log.Error(
			errUnmarsConfig,
			zap.String(infFile, c.tableName),
			zap.Error(err))

		return Config{}, errors.Wrap(err, errUnmarsConfig)
	}

	if len(resMap) == 0 {
		return DefaultConfig(), nil
	}

	var mc Config

	if err := json.Unmarshal(
		[]byte(resMap[dynamoDBTableConfigName]), &mc); err != nil {
		c.log.Error(
			errUnmarsConfig,
			zap.String(infFile, c.tableName),
			zap.Error(err))

		return Config{}, errors.Wrap(err, errUnmarsConfig)
	}

	c.log.Info(infConfigLoaded, zap.String(infConfig, mc.String()))

	return mc, nil
}

// AWSDynamoConfigWrite writes the micro controller configuration
// from AWS dynamodb
type AWSDynamoConfigWrite struct {
	log       *zap.Logger
	hub       Hub
	config    config.Config
	tableName string
	client    *dynamodb.Client
}

// NewAWSConfigWrite creates the micro controller configuration
// from AWS dynamodb
func NewAWSConfigWrite(
	log *zap.Logger,
	hub Hub,
	config config.Config,
	cfg aws.Config,
	tableName string) *AWSDynamoConfigWrite {
	//
	return &AWSDynamoConfigWrite{
		log:       log,
		hub:       hub,
		config:    config,
		tableName: tableName,
		client:    dynamodb.NewFromConfig(cfg),
	}
}

func (c AWSDynamoConfigWrite) Save(data Config) error {
	c.log.Info(
		infSavingConfig,
		zap.String(infConfig, data.String()),
		zap.String(infFile, c.tableName))

	conf, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, strings.Concat(errMarshallConfig, c.tableName))
	}

	_, err = c.client.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{
			TableName: aws.String(c.tableName),
			Item: map[string]types.AttributeValue{
				dynamoDBTableKeyName: &types.AttributeValueMemberS{
					Value: dynamoDBTableKeyValue},
				dynamoDBTableConfigName: &types.AttributeValueMemberS{
					Value: string(conf)},
			},
		})

	if err != nil {
		return errors.Wrap(err, strings.Concat(errSaveConfig, c.tableName))
	}

	notifyHub(c.config, data, c.hub)

	return nil
}

func notifyHub(c config.Config, data Config, h Hub) {
	h.Config(iot.DeviceConfig{
		CollectMetricsTime:   c.CollectMetricsTime,
		WakeUpTime:           data.Wakeup,
		Buffer:               data.Buffer,
		IniSendTime:          data.IniSendTime,
		EndSendTime:          data.EndSendTime,
		CalibratingORP:       data.CalibratingORP,
		TargetORP:            data.TargetORP,
		CalibrationORP:       data.CalibrationORP,
		StabilizationTimeORP: data.StabilizationTimeORP,
	})
}
