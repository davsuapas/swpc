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

package ai

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const (
	errCreateTempFile  = "Error creating temporal file"
	errPredict         = "Error calling predict.sh"
	errReadErrorResult = "Error reading error file"
	errModelExist      = "The prediction model does not exist. File: "
)

const (
	infCommand = "Executing predict.sh"
)

const (
	wqModelFile = "ai/model_wq"
	clModelFile = "ai/model_cl"
)

// Predicter generates predictions
type Predicter interface {
	// predict predicts the chlorine and water quality
	Predict(temp string, ph string, orp string) (string, error)
}

// Prediction is used as interface
type Prediction struct {
	Log *zap.Logger
}

// WQ predcits the water quality
// The function calls a python script
// The models
func (p *Prediction) Predict(
	temp string,
	ph string,
	orp string) (string, error) {
	//
	errorFile, err := os.CreateTemp("", "error_predict")
	if err != nil {
		return "", errors.Wrap(err, errCreateTempFile)
	}

	errorFile.Close()

	defer os.Remove(errorFile.Name())

	if _, err := os.Stat(wqModelFile); errors.Is(err, os.ErrNotExist) {
		return "", errors.Wrap(err, strings.Concat(errModelExist, wqModelFile))
	}

	if _, err := os.Stat(clModelFile); errors.Is(err, os.ErrNotExist) {
		return "", errors.Wrap(err, strings.Concat(errModelExist, clModelFile))
	}

	var stdoutBuf bytes.Buffer

	p.Log.Info(
		infCommand,
		zap.String("mwq", wqModelFile),
		zap.String("mcl", clModelFile),
		zap.String("e", errorFile.Name()),
		zap.String("temp", temp),
		zap.String("ph", ph),
		zap.String("ph", orp))

	cmd := exec.Command(
		"ai/predict.sh",
		"-mwq", wqModelFile,
		"-mcl", clModelFile,
		"-e", errorFile.Name(),
		"-temp", temp,
		"-ph", ph,
		"-orp", orp,
	) // #nosec G204

	cmd.Stdout = &stdoutBuf

	if errCommand := cmd.Run(); errCommand != nil {
		result, err := os.ReadFile(errorFile.Name())
		if err != nil {
			return "", errors.Wrap(errCommand, errReadErrorResult)
		}

		return "", errors.Wrap(errCommand, string(result))
	}

	return stdoutBuf.String(), nil
}
