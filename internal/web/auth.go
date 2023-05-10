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

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/auth"
	"go.uber.org/zap"
)

const (
	errTkInvalid     = "Auth -> Token invalid"
	errStateInvalid  = "Auth -> Auth state invalid"
	errRevokeTk      = "Auth -> Revoke token caused by logoff"
	errGetTk         = "Auth -> Getting auth token from web request"
	errGetWSClientID = "Auth -> Getting web socket client id from web request"
)

const (
	infLogin  = "Auth -> Get token for login"
	infLogoff = "Auth -> Remove token for logoff"
)

const (
	AuthHeaderName = "Authorization"
	AuthCheckName  = "IsAuth"
)

const (
	RedirectLoginOk   = "/"
	RedirectLogin     = "/auth/login"
	RedirectLogout    = "/auth/logout"
	RedirectErrorAuth = "/auth/error"
)

type OAuth2 interface {
	Token(params auth.OA2TokenInput) (*jwt.Token, error)
	RevokeToken(params auth.OA2RevokeTokenInput) error
}

type Auth interface {
	Login(ctx echo.Context) error
	Logout(ctx echo.Context) error
}

// AuthFlow manages authentication of oauth2 supported providers
type AuthFlow struct {
	Log     *zap.Logger
	Service OAuth2
	Hub     Hub
	Config  config.Config
}

// Login gets a token from the provider
func (o *AuthFlow) Login(ctx echo.Context) error {
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")

	o.Log.Info(infLogin, zap.String("Code", code), zap.String("State", state))

	if _, err := auth.DecodeState([]byte(o.Config.SecretKey), state); err != nil {
		o.Log.Error(errStateInvalid, zap.Error(err))

		return ctx.Redirect(http.StatusFound, RedirectErrorAuth)
	}

	param := auth.OA2TokenInput{
		URL:         o.Config.Auth.TokenURL,
		Code:        code,
		RedirectURI: o.Config.AuthRedirectURI(RedirectLogin),
	}

	token, err := o.Service.Token(param)
	if err != nil || !token.Valid {
		o.Log.Error(errTkInvalid, zap.Error(err))

		return ctx.Redirect(http.StatusFound, RedirectErrorAuth)
	}

	// Save the security token in the cookies
	// MaxAge is the same time than token expiration except expiration for jwt token.
	// The expiration of the cookie with the token is kept 5 minutes longer than
	// the internal expiration of the token, in case a new request is made from the browser.
	// In this way, the server will return permission denied instead of bad request
	// (this case would be because when the cookie expires the request would come without a token).
	expiration := time.Now().Add(time.Duration(o.Config.Web.SessionExpiration) * time.Minute)
	cookie := cookies(AuthHeaderName, token.Raw, expiration.Add(5*time.Minute))
	ctx.SetCookie(cookie)

	cookie = cookieAuthCheckName(expiration)
	ctx.SetCookie(cookie)

	// ID for hub client
	cookie = cookieWSClientIDName(expiration)
	ctx.SetCookie(cookie)

	return ctx.Redirect(http.StatusFound, RedirectLoginOk)
}

// Logout revokes token, unregisters client into hub and delete cookies
func (o *AuthFlow) Logout(ctx echo.Context) error {
	o.Log.Info(infLogoff)

	if !unregisterHub(ctx, o.Log, o.Hub) {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	token, err := ctx.Cookie(AuthHeaderName)
	if err != nil {
		o.Log.Error(errGetTk, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	params := auth.OA2RevokeTokenInput{
		URL:   o.Config.Auth.RevokeTokenURL,
		Token: token.Value,
	}

	if err := o.Service.RevokeToken(params); err != nil {
		o.Log.Error(errRevokeTk, zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	cookie := cookies(AuthHeaderName, "", time.Time{})
	cookie.MaxAge = 0 // Remove cookie
	ctx.SetCookie(cookie)

	cookie = removeCookieAuthCheckName()
	ctx.SetCookie(cookie)

	return ctx.NoContent(http.StatusOK)
}

// AuthFlowDev manages authentication only for develoment
type AuthFlowDev struct {
	Log  *zap.Logger
	Hub  Hub
	Webc config.Web
}

// Login enables insecure tokenless access
func (o *AuthFlowDev) Login(ctx echo.Context) error {
	o.Log.Info(infLogin)

	expiration := time.Now().Add(time.Duration(o.Webc.SessionExpiration) * time.Minute)

	cookie := cookieAuthCheckName(expiration)
	ctx.SetCookie(cookie)

	// ID for hub client
	cookie = cookieWSClientIDName(expiration)
	ctx.SetCookie(cookie)

	return ctx.Redirect(http.StatusFound, RedirectLoginOk)
}

// Logout unregisters client into hub and delete cookies
func (o *AuthFlowDev) Logout(ctx echo.Context) error {
	o.Log.Info(infLogoff)

	if !unregisterHub(ctx, o.Log, o.Hub) {
		return ctx.Redirect(http.StatusFound, RedirectErrorAuth)
	}

	cookie := removeCookieAuthCheckName()
	ctx.SetCookie(cookie)

	return ctx.Redirect(http.StatusFound, RedirectLoginOk)
}

func unregisterHub(ctx echo.Context, log *zap.Logger, hub Hub) bool {
	id, err := ctx.Cookie(WSClientIDName)
	if err != nil {
		log.Error(errGetWSClientID, zap.Error(err))

		return false
	}

	hub.Unregister(id.Value)

	return true
}

func cookieAuthCheckName(expiration time.Time) *http.Cookie {
	cookie := cookies(AuthCheckName, "true", expiration)
	cookie.HttpOnly = false

	return cookie
}

func cookieWSClientIDName(expiration time.Time) *http.Cookie {
	cookie := cookies(WSClientIDName, xid.New().String(), expiration.Add(5*time.Minute))

	return cookie
}

func removeCookieAuthCheckName() *http.Cookie {
	cookie := cookies(AuthCheckName, "", time.Time{})
	cookie.HttpOnly = false
	cookie.MaxAge = 0 // Remove cookie

	return cookie
}

func cookies(name string, value string, expiration time.Time) *http.Cookie {
	cookie := &http.Cookie{}
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"
	cookie.Expires = expiration
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode

	return cookie
}
