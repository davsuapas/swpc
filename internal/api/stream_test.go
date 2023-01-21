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

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/swpoolcontroller/internal/api"
	"github.com/swpoolcontroller/internal/micro"
	"github.com/swpoolcontroller/internal/micro/mocks"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

func TestStream_Actions(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/actions/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	mhub := mocks.NewHub(t)

	mhub.On("Status", mock.AnythingOfType("chan sockets.Status")).Run(func(args mock.Arguments) {
		if s, ok := args.Get(0).(chan sockets.Status); ok {
			go func() {
				s <- sockets.Closed
			}()
		}
	})

	c := api.NewStream(&micro.Controller{
		Log:                zap.NewExample(),
		Hub:                mhub,
		Config:             micro.ConfigDefault(),
		CheckTransTime:     10,
		CollectMetricsTime: 20,
	})

	_ = c.Actions(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(
		t,
		"{\"WakeUpTime\":30,\"CheckTransTime\":10,\"CollectMetricsTime\":20,\"Buffer\":3,\"Action\":0}\n",
		rec.Body.String())

	mhub.AssertExpectations(t)
}

func TestStream_Download(t *testing.T) {
	t.Parallel()

	metrics := "1,2,3"

	e := echo.New()
	body := strings.NewReader(metrics)
	req := httptest.NewRequest(http.MethodGet, "/download/", body)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	mhub := mocks.NewHub(t)

	mhub.On("Send", metrics)
	mhub.On("Status", mock.AnythingOfType("chan sockets.Status")).Run(func(args mock.Arguments) {
		if s, ok := args.Get(0).(chan sockets.Status); ok {
			go func() {
				s <- sockets.Closed
			}()
		}
	})

	c := api.NewStream(&micro.Controller{
		Log:                zap.NewExample(),
		Hub:                mhub,
		Config:             micro.ConfigDefault(),
		CheckTransTime:     10,
		CollectMetricsTime: 20,
	})

	_ = c.Download(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(
		t,
		"{\"WakeUpTime\":30,\"CheckTransTime\":10,\"CollectMetricsTime\":20,\"Buffer\":3,\"Action\":0}\n",
		rec.Body.String())

	mhub.AssertExpectations(t)
}
