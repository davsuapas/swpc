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
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/ai"
	"go.uber.org/zap"
)

const (
	errGettingMetrics = "Getting the metrics of the request body to predict"
	errPredict        = "Getting the water quality and chlorine prediction"
	errMarshalPreds   = "Marshal json predictions"
	errToken          = "The token is bad"
)

const (
	infMetrics = "Metrics to predict water quality and chlorine"
)

type metrics struct {
	Temp string `json:"temp"`
	PH   string `json:"ph"`
	ORP  string `json:"orp"`
}

type predictions struct {
	WQ string `json:"wq"`
	CL string `json:"cl"`
}

type PredictionWeb struct {
	Preder ai.Predicter
	Log    *zap.Logger
}

func (p *PredictionWeb) Predict(ctx echo.Context) error {
	var metrics metrics

	if err := ctx.Bind(&metrics); err != nil {
		p.Log.Error(errGettingMetrics, zap.Error(err))

		return ctx.NoContent(http.StatusBadRequest)
	}

	p.Log.Info(
		infMetrics,
		zap.String("Temp", metrics.Temp),
		zap.String("PH", metrics.PH),
		zap.String("ORP", metrics.ORP))

	preds, err := p.Preder.Predict(
		metrics.Temp,
		metrics.PH,
		metrics.ORP)

	if err != nil {
		p.Log.Error(errPredict, zap.Error(err))

		if errors.Is(err, os.ErrNotExist) {
			return ctx.NoContent(http.StatusNotFound)
		}

		return ctx.NoContent(http.StatusInternalServerError)
	}

	tokenPred := strings.Split(preds, ";")

	if len(tokenPred) != 2 {
		p.Log.Info(errToken, zap.String("token", preds))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := predictions{
		WQ: tokenPred[0],
		CL: tokenPred[1],
	}

	p.Log.Info(
		infMetrics,
		zap.String("Water Quality", res.WQ),
		zap.String("Chlorine", res.CL))

	return ctx.JSON(http.StatusOK, res)
}
