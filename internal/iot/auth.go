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

package iot

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"go.uber.org/zap"
)

const ClientIDName = "client_id"

const (
	errBadID = "OAuth.Token.The secretID is bad"
	errSign  = "OAuth.Token.Error signing token"
)

const dbgGetTk = "OAuth.Token.Getting token"

// Auth controllers the access of the API
type Auth struct {
	log *zap.Logger
	ac  config.API
}

func NewAuth(log *zap.Logger, ac config.API) *Auth {
	return &Auth{
		log: log,
		ac:  ac,
	}
}

// Token gets security token
func (o *Auth) Token(ctx echo.Context) error {
	o.log.Debug(dbgGetTk)

	sID := ctx.Param(ClientIDName)

	if sID != o.ac.ClientID {
		o.log.Error(errBadID)

		return ctx.NoContent(http.StatusUnauthorized)
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(
			time.Now().Add(time.Duration(5) * time.Minute)),
	}

	// Create token with claims
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	token, err := t.SignedString([]byte(o.ac.TokenSecretKey))
	if err != nil {
		o.log.With(zap.Error(err)).Error(errSign, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.String(http.StatusOK, token)
}
