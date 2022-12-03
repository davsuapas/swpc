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

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/pkg/sockets"
	"go.uber.org/zap"
)

// WS register sockets
type WS struct {
	log      *zap.Logger
	hub      *sockets.Hub
	sessionc config.WebConfig
	upgrader websocket.Upgrader
}

// NewWS builds WS service
func NewWS(log *zap.Logger, sessionc config.WebConfig, hub *sockets.Hub) *WS {
	return &WS{
		log:      log,
		hub:      hub,
		sessionc: sessionc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Register registers sockets from web request
func (w *WS) Register(ctx echo.Context) error {
	ws, err := w.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		w.log.Error("WS. Generating socket from web request", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	sess, err := ctx.Cookie(TokenName)
	if err != nil {
		w.log.Error("WS. Getting auth token from web request", zap.Error(err))

		return ctx.NoContent(http.StatusInternalServerError)
	}

	w.hub.Register(sockets.NewClient(sess.Value, ws, time.Duration(w.sessionc.SessionExpiration)*time.Minute))

	return ctx.NoContent(http.StatusOK)
}
