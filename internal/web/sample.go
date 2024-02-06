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

package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/ai"
	"go.uber.org/zap"
)

// Web error
const (
	errGettingSample = "Getting the sample of the request body"
	errSaveSample    = "Saving sample data in repository"
)

const (
	infNotImplementedRepo = "You are using a sample repo " +
		"that does not perform any actions. Use a cloud provider"
)

// SampleWeb saves the samples in the repo according to repository type
type SampleWeb struct {
	Log  *zap.Logger
	Repo ai.SampleRepo
}

// Save saves the samples in the repo
func (s *SampleWeb) Save(ctx echo.Context) error {
	var sample ai.SampleData

	if err := ctx.Bind(&sample); err != nil {
		s.Log.Error(errGettingSample, zap.Error(err))

		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := s.Repo.Save(sample); err != nil {
		s.Log.Error(errSaveSample, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

// SampleDummyRepo is a not implemented repo
type SampleDummyRepo struct {
	Log *zap.Logger
}

func (s *SampleDummyRepo) Save(d ai.SampleData) error {
	s.Log.Info(infNotImplementedRepo, zap.String("sample", d.String()))

	return nil
}
