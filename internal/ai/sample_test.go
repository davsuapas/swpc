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

package ai_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swpoolcontroller/internal/ai"
	"go.uber.org/zap"
)

func TestSampleFileRepo_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sample   ai.SampleData
		fileName string
		wantErr  bool
	}{
		{
			name: "Save sample data to file successfully",
			sample: ai.SampleData{
				Temp:     "12.1",
				PH:       "2.1",
				ORP:      "-12.1",
				Quality:  "0",
				Chlorine: "-23.1",
			},
			fileName: "./testr/sample.csv",
			wantErr:  false,
		},
		{
			name: "Save sample data to file with error",
			sample: ai.SampleData{
				Temp:     "12.1",
				PH:       "2.1",
				ORP:      "-12.1",
				Quality:  "0",
				Chlorine: "-23.1",
			},
			fileName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := zap.NewExample()
			repo := ai.SampleFileRepo{
				Log:      log,
				FileName: tt.fileName,
			}

			err := repo.Save(tt.sample)
			if err == nil {
				if err := os.Remove(tt.fileName); err != nil {
					require.Error(t, err, "Removing sample data file created")
				}
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
