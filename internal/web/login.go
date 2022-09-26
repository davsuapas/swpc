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

package web

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/crypto"
	pcrypto "github.com/swpoolcontroller/pkg/crypto"
	"go.uber.org/zap"
)

const (
	tokenExpiration = 10
	TokenName       = "Authorization"
)

// Login controllers the access of the user
type Login struct {
	log   *zap.Logger
	users users
}

func NewLogin(log *zap.Logger) *Login {
	return &Login{
		log:   log,
		users: newUsersInMemory(),
	}
}

// Submit validates the user, generates token and save session
func (l *Login) Submit(c echo.Context) error {
	email := c.FormValue("email")
	pass := c.FormValue("password")

	// Validate user
	userPass, ok := l.users.get(email)
	if !ok {
		l.log.Error("The user not found", zap.String("user", email))
		return c.NoContent(http.StatusUnauthorized)
	}

	ep, err := pcrypto.Encrypt(pass, crypto.Key)
	if err != nil {
		l.log.With(zap.Error(err)).Error("Error encrypt the pass")
		return c.NoContent(http.StatusUnauthorized)
	}
	if ep != userPass {
		l.log.Error("The pass is bad", zap.String("user", email))
		return c.NoContent(http.StatusUnauthorized)
	}

	// Security
	token, err := l.securityToken(c, email)
	if err != nil {
		return err
	}

	// Save the security token in the cookies
	// MaxAge is the same time than token expiration
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * tokenExpiration,
		HttpOnly: true,
	}
	sess.Values[TokenName] = token
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		l.log.With(zap.Error(err)).Error("Error saving the session")
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// securityToken generates the jwt security token
func (l *Login) securityToken(c echo.Context, email string) (string, error) {
	// Set custom claims
	claims := &JWTCustomClaims{
		email,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * tokenExpiration).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(crypto.Key))
	if err != nil {
		l.log.With(zap.Error(err)).Error("Error signing token", zap.Error(err))
		return "", c.NoContent(http.StatusInternalServerError)
	}
	return t, nil
}

// JWTCustomClaims are custom claims extending default ones.
type JWTCustomClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Users controllers the user of app in memory
type users struct {
	user map[string]string
}

func newUsersInMemory() users {
	return users{
		user: map[string]string{
			"dav.sua.pas@gmail.com": "RCrkRDBG6cc=",
		},
	}
}

// get gets the user. If the user is not found return ("", false)
func (u *users) get(user string) (string, bool) {
	value, ok := u.user[user]
	return value, ok
}
