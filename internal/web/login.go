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
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/crypto"
	pcrypto "github.com/swpoolcontroller/pkg/crypto"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

const (
	TokenName  = "Authorization"
	authCookie = "IsAuth"
)

const (
	errGettingToken = "Logoff. Getting auth token from web request"
	errUserNotFound = "Login. The user not found"
	errEncrypt      = "Login. Error encrypt the pass"
	errBadKey       = "Login. The pass is bad"
	errSign         = "Login. Error signing token"
)

const (
	infLoginRequest  = "Login request"
	infLogoffRequest = "Logoff request"
)

// Login controllers the access of the user
type Login struct {
	log   *zap.Logger
	hub   *sockets.Hub
	webc  config.WebConfig
	users users
}

func NewLogin(log *zap.Logger, wc config.WebConfig, hub *sockets.Hub) *Login {
	return &Login{
		log:   log,
		webc:  wc,
		users: newUsersInMemory(),
		hub:   hub,
	}
}

// Logoff removes cookies and it un-registers the socket
func (l *Login) Logoff(ctx echo.Context) error {
	l.log.Info(infLoginRequest)

	cookie := cookies(TokenName, "", true, time.Time{})
	cookie.MaxAge = 0 // Remove cookie
	ctx.SetCookie(cookie)
	cookie = cookies(authCookie, "", false, time.Time{})
	cookie.MaxAge = 0 // Remove cookie
	ctx.SetCookie(cookie)

	sess, err := ctx.Cookie(TokenName)
	if err != nil {
		l.log.Error(errGettingToken, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.hub.Unregister(sess.Value)

	return ctx.NoContent(http.StatusOK)
}

// Submit validates the user, generates token and save session
func (l *Login) Submit(ctx echo.Context) error {
	email := ctx.FormValue("email")
	pass := ctx.FormValue("password")

	l.log.Info(infLoginRequest, zap.String("user", email))

	// Validate user
	userPass, ok := l.users.get(email)
	if !ok {
		l.log.Error(errUserNotFound, zap.String("user", email))

		return ctx.NoContent(http.StatusUnauthorized)
	}

	passEncrypt, err := pcrypto.Encrypt(pass, crypto.Key)
	if err != nil {
		l.log.With(zap.Error(err)).Error(errEncrypt)

		return ctx.NoContent(http.StatusUnauthorized)
	}

	if passEncrypt != userPass {
		l.log.Error(errBadKey, zap.String("user", email))

		return ctx.NoContent(http.StatusUnauthorized)
	}

	// Security
	token, err := l.securityToken(ctx, email)
	if err != nil {
		return err
	}

	// Save the security token in the cookies
	// MaxAge is the same time than token expiration except expiration for jwt token.
	// The expiration of the cookie with the token is kept 5 minutes longer than
	// the internal expiration of the token, in case a new request is made from the browser.
	// In this way, the server will return permission denied instead of bad request
	// (this case would be because when the cookie expires the request would come without a token).
	expiration := time.Now().Add(time.Duration(l.webc.SessionExpiration) * time.Minute)
	cookie := cookies(TokenName, token, true, expiration.Add(5*time.Minute))
	ctx.SetCookie(cookie)
	cookie = cookies(authCookie, "true", false, expiration)
	ctx.SetCookie(cookie)

	return ctx.NoContent(http.StatusOK)
}

func cookies(name string, value string, httpOnly bool, expiration time.Time) *http.Cookie {
	cookie := &http.Cookie{}
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"
	cookie.Expires = expiration
	cookie.HttpOnly = httpOnly
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode

	return cookie
}

// securityToken generates the jwt security token
func (l *Login) securityToken(ctx echo.Context, email string) (string, error) {
	// Set custom claims
	claims := &JWTCustomClaims{
		email,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(l.webc.SessionExpiration) * time.Minute).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	tsigned, err := token.SignedString([]byte(crypto.Key))
	if err != nil {
		l.log.With(zap.Error(err)).Error(errSign, zap.Error(err))

		return "", ctx.NoContent(http.StatusInternalServerError)
	}

	return tsigned, nil
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
