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

package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swpoolcontroller/internal/ai"
	"github.com/swpoolcontroller/internal/ai/mocks"
	"github.com/swpoolcontroller/internal/web"
	"go.uber.org/zap"
)

func TestSample_Save(t *testing.T) {
	t.Parallel()

	type mSampleRepo struct {
		apply bool
		s     ai.SampleData
		err   error
	}

	tests := []struct {
		name        string
		argBody     string
		mSampleRepo mSampleRepo
		resStatus   int
	}{
		{
			name: "Save sample data. StatusOk",
			argBody: `{"temp": 12.1, "ph": 2.1, "orp": -12.1,
			 "quality": 0, "chlorine": -23.1}`,
			mSampleRepo: mSampleRepo{
				apply: true,
				s: ai.SampleData{
					Temp:     12.1,
					PH:       2.1,
					ORP:      -12.1,
					Quality:  0,
					Chlorine: -23.1,
				},
				err: nil,
			},
			resStatus: http.StatusOK,
		},
		{
			name:    "Save sample data with body bad. StatusBadRequest",
			argBody: `{"temp": `,
			mSampleRepo: mSampleRepo{
				apply: false,
				s:     ai.SampleData{},
				err:   nil,
			},
			resStatus: http.StatusBadRequest,
		},
		{
			name: "Save sample data with failed repo. StatusInternalServerError",
			argBody: `{"temp": 12, "ph": 7, "orp": -45,
			 "quality": 1, "chlorine": 23}`,
			mSampleRepo: mSampleRepo{
				apply: true,
				s: ai.SampleData{
					Temp:     12,
					PH:       7,
					ORP:      -45,
					Quality:  1,
					Chlorine: 23,
				},
				err: errors.New("failed repo"),
			},
			resStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mrepo := mocks.NewSampleRepo(t)

			s := web.SampleWeb{
				Log:  zap.NewExample(),
				Repo: mrepo,
			}

			e := echo.New()
			body := strings.NewReader(tt.argBody)
			req := httptest.NewRequest(http.MethodPost, "/sample", body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			if tt.mSampleRepo.apply {
				mrepo.On("Save", tt.mSampleRepo.s).Return(tt.mSampleRepo.err)
			}

			_ = s.Save(ctx)

			assert.Equal(t, tt.resStatus, rec.Code)

			mrepo.AssertExpectations(t)
		})
	}
}

func TestSampleDummyRepo_Save(t *testing.T) {
	t.Parallel()

	s := web.SampleDummyRepo{
		Log: zap.NewExample(),
	}

	err := s.Save(ai.SampleData{})

	require.NoError(t, err)
}
