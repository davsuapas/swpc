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
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal/ai/mocks"
	"github.com/swpoolcontroller/internal/web"
	"go.uber.org/zap"
)

func TestPrediction_Predict(t *testing.T) {
	t.Parallel()

	type mPreds struct {
		apply bool
		err   error
		token string
	}

	tests := []struct {
		name      string
		argBody   string
		mPreds    mPreds
		resStatus int
	}{
		{
			name:    "Predict with metrics. It should return the predictions values",
			argBody: `{"temp":"32","ph":"8.9","orp":"-465"}`,
			mPreds: mPreds{
				apply: true,
				err:   nil,
				token: "regular;123.23",
			},
			resStatus: http.StatusOK,
		},
		{
			name:    "Predict with body bad. StatusBadRequest",
			argBody: `{"tem-465"}`,
			mPreds: mPreds{
				apply: false,
				err:   nil,
				token: "regular;123.23",
			},
			resStatus: http.StatusBadRequest,
		},
		{
			name:    "Predict with bad token. StatusInternalServerError",
			argBody: `{"temp":"32","ph":"8.9","orp":"-465"}`,
			mPreds: mPreds{
				apply: true,
				err:   nil,
				token: "",
			},
			resStatus: http.StatusInternalServerError,
		},
		{
			name:    "Predict with error executing script. StatusInternalServerError",
			argBody: `{"temp":"32","ph":"8.9","orp":"-465"}`,
			mPreds: mPreds{
				apply: true,
				err:   errors.New("Error executing python script"),
			},
			resStatus: http.StatusInternalServerError,
		},
		{
			name:    "Predict. Model file not exist. StatusNotFound",
			argBody: `{"temp":"32","ph":"8.9","orp":"-465"}`,
			mPreds: mPreds{
				apply: true,
				err:   errors.Wrap(os.ErrNotExist, "Model not exist"),
			},
			resStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mpred := mocks.NewPredicter(t)

			s := web.PredictionWeb{
				Log:    zap.NewExample(),
				Preder: mpred,
			}

			e := echo.New()
			body := strings.NewReader(tt.argBody)
			req := httptest.NewRequest(http.MethodPost, "/predict", body)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)

			if tt.mPreds.apply {
				mpred.On(
					"Predict", "32", "8.9", "-465").Return(
					tt.mPreds.token, tt.mPreds.err)
			}

			_ = s.Predict(ctx)

			assert.Equal(t, tt.resStatus, rec.Code)

			if rec.Code == http.StatusOK {
				assert.JSONEq(
					t,
					"{\"wq\":\"regular\",\"cl\":\"123.23\"}\n",
					rec.Body.String())
			}

			mpred.AssertExpectations(t)
		})
	}
}
