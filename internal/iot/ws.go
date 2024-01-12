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

package iot

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/pkg/iot"
	"go.uber.org/zap"
)

const (
	errGenSocket = "WS Device. Generating socket from device request"
)

const (
	infRegisterDevice = "Registering iot device"
)

// WS register sockets
type WS struct {
	log      *zap.Logger
	hub      Hub
	upgrader websocket.Upgrader
}

// NewWS builds WS service
func NewWS(log *zap.Logger, hub Hub) *WS {
	return &WS{
		log: log,
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Register registers sockets from device request
func (w *WS) Register(ctx echo.Context) error {
	ws, err := w.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		w.log.Error(errGenSocket, zap.Error(err))

		// Upgrade update the response. No need to return the error
		return nil
	}

	deviceID := ctx.Request().Header.Get("id")

	w.log.Info(infRegisterDevice, zap.String("deviceID", deviceID))

	w.hub.RegisterDevice(
		iot.Device{
			ID:         deviceID,
			Connection: ws,
		})

	// Upgrade update the response. No need to return the error
	return ctx.NoContent(http.StatusOK)
}
