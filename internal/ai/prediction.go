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
)

const (
	errCreateTempFile  = "Error creating temporal file"
	errPredict         = "Error calling predict.sh"
	errReadErrorResult = "Error reading error file"
)

// Predicter generates predictions
type Predicter interface {
	// predict predicts the chlorine and water quality
	Predict(temp string, ph string, orp string) (string, error)
}

// PredictFunc is used as interface
type PredictFunc struct {
}

// WQ predcits the water quality
// The function calls a python script
// The models
func (p *PredictFunc) Predict(
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

	var stdoutBuf bytes.Buffer

	cmd := exec.Command(
		"ai/predict.sh",
		"-mwq", "ai/model_wq",
		"-mcl", "ai/model_cl",
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
