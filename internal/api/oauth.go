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

package api

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/crypto"
	"go.uber.org/zap"
)

const (
	SName = "secretID"
	sid   = "sw3kf$fekdy56dfh"
)

// OAuth controllers the access of the API
type OAuth struct {
	log *zap.Logger
	ac  config.APIConfig
}

func NewOAuth(log *zap.Logger, ac config.APIConfig) *OAuth {
	return &OAuth{
		log: log,
		ac:  ac,
	}
}

// Token gets security token
func (o *OAuth) Token(ctx echo.Context) error {
	sID := ctx.Param(SName)

	if sID != sid {
		o.log.Error("OAuth.Token. The scretID is bad")

		return ctx.NoContent(http.StatusUnauthorized)
	}

	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Duration(o.ac.SessionExpiration) * time.Minute).Unix(),
	}

	// Create token with claims
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	token, err := t.SignedString([]byte(crypto.Key))
	if err != nil {
		o.log.With(zap.Error(err)).Error("OAuth.Token. Error signing token", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.String(http.StatusOK, token)
}
