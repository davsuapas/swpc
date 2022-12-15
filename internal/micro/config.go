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

package micro

import (
	"encoding/json"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const fileName = "micro-config.dat"

const (
	errReadConfig     = "Reading the configuration of the micro controller from config file"
	errUnmarsConfig   = "Unmarshalling the configuration of the micro controller from config file"
	errMarshallConfig = "Marshalling the configuration of the micro controller: "
	errSaveConfig     = "Saving the configuration for the micro controller: "
)

type Config struct {
	// IniSendTime is the range for initiating metric sends
	IniSendTime string `json:"iniSendTime"`
	// EndSendTime is the range for ending metric sends
	EndSendTime string `json:"endSendTime"`
	// Wakeup is how often in minutes the micro-controller wakes up to check for sending
	Wakeup uint8 `json:"wakeup"`
	// Buffer is the buffer in seconds to store metrics int the micro-controller
	Buffer uint8 `json:"buffer"`
}

func configDefault() Config {
	return Config{
		IniSendTime: "11:00",
		EndSendTime: "12:00",
		Wakeup:      30,
		Buffer:      5,
	}
}

// ConfigRead reads the micro controller configuration
type ConfigRead struct {
	Log      *zap.Logger
	DataPath string
}

// ConfigWrite writes the micro controller configuration
type ConfigWrite struct {
	Log      *zap.Logger
	MControl *Controller
	DataPath string
}

// Read reads the configuration saved in disk. If the file not exists returns config default
func (c ConfigRead) Read() (Config, error) {
	file := path.Join(c.DataPath, fileName)

	data, err := os.ReadFile(file)
	if err != nil {
		c.Log.Error(errReadConfig, zap.String("file", file), zap.Error(err))

		if errors.Is(err, os.ErrNotExist) {
			return configDefault(), nil
		}

		return Config{}, nil
	}

	var mc Config

	if err := json.Unmarshal(data, &mc); err != nil {
		c.Log.Error(errUnmarsConfig, zap.String("file", file), zap.Error(err))

		return Config{}, nil
	}

	return mc, nil
}

func (c ConfigWrite) Save(data Config) error {
	file := path.Join(c.DataPath, fileName)

	conf, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, strings.Concat(errMarshallConfig, file))
	}

	if err := os.WriteFile(file, conf, os.FileMode(0664)); err != nil {
		return errors.Wrap(err, strings.Concat(errSaveConfig, file))
	}

	// it updates the controller with new configuration
	c.MControl.SetConfig(data)

	return nil
}
