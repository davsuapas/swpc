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

package config

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const fileName = "micro-config.dat"

// MicroConfig manages the micro controller configuration
type MicroConfig struct {
	Log      *zap.Logger
	DataPath string
}

// Read reads the configuration saved in disk
func (c MicroConfig) Read() ([]byte, error) {
	file := path.Join(c.DataPath, fileName)
	data, err := os.ReadFile(file)
	if err != nil {
		return []byte{}, errors.Wrap(err, strings.Concat("Reading the configuration for the micro controller: ", file))
	}
	return data, nil
}

func (c MicroConfig) Save(data []byte) error {
	file := path.Join(c.DataPath, fileName)
	if err := os.WriteFile(file, data, os.FileMode(0664)); err != nil {
		return errors.Wrap(err, strings.Concat("Saving the configuration for the micro controller: ", file))
	}
	return nil
}
