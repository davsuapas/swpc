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

package internal

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/swpoolcontroller/internal/config"
	"github.com/swpoolcontroller/internal/iot"
	"github.com/swpoolcontroller/internal/web"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const (
	errShutdownServer = "Shutting down web server"
)

const (
	infStartingServer    = "Starting the swimming pool controller server ..."
	infStartHub          = "Starting the hub"
	infStoppingWebServer = "The web server is stopping ..."
	infStoppedWebServer  = "The web server has been stopped"
	infStoppingHub       = "The hub is stopping ..."
	infStoppedServer     = "The server has been stopped"
)

type Server struct {
	factory *Factory
	quit    chan os.Signal
}

func NewServer(factory *Factory) *Server {
	return &Server{
		factory: factory,
		quit:    make(chan os.Signal, 1),
	}
}

// Kill forces to kill the process
func (s *Server) Kill() {
	s.quit <- os.Kill
}

// Start starts the graceful http server and services
func (s *Server) Start() error {
	s.factory.Log.Info(infStartingServer)

	// Start server
	go func() {
		s.factory.Hubt.Register()

		address := strings.Concat(
			s.factory.Config.Server.Internal.Host, ":",
			strconv.Itoa(s.factory.Config.Internal.Port))

		s.factory.Log.Info(infStartHub, zap.String("Address", address))
		s.factory.Hub.Run()

		if err := s.factory.Webs.Start(address); err != nil {
			s.factory.Log.Info(infStoppedWebServer + ". " + err.Error())
		}
	}()

	signal.Notify(s.quit, os.Interrupt)
	<-s.quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.factory.Log.Info(infStoppingWebServer)

	if err := s.factory.Webs.Shutdown(ctx); err != nil {
		erre := errors.Wrap(err, errShutdownServer)
		s.factory.Log.Error(erre.Error())

		return errors.Wrap(erre, errShutdownServer)
	}

	s.factory.Log.Info(infStoppingHub)
	s.factory.Hub.Stop()

	s.factory.Log.Info(infStoppedServer)

	return nil
}

// Middleware configure security and behaviour of http
func (s *Server) Middleware() {
	s.factory.Webs.Use(middleware.Recover())

	// SPA web
	s.factory.Webs.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "public",
		Index:  "index.html",
		Browse: false,
		HTML5:  true,
	}))
}

// Route sets the router of app so web as api
func (s *Server) Route() {
	// Public
	wapp := s.factory.Webs.Group("/app")
	wapp.GET("/config", s.factory.WebHandler.AppConfig.Load)

	wa := s.factory.Webs.Group("/auth")
	wa.GET("/login", s.factory.WebHandler.Auth.Login)
	wa.GET("/logout", s.factory.WebHandler.Auth.Logout)
	wa.GET(strings.Concat(
		"/token/:",
		iot.ClientIDName),
		s.factory.APIHandler.Auth.Token)

	// API Restricted by JWT

	// Web
	wapi := s.factory.Webs.Group("/api/web")

	if s.factory.Config.Auth.Provider == config.AuthProviderOauth2 {
		config := echojwt.Config{
			KeyFunc:     s.factory.JWT.GetKey,
			TokenLookup: strings.Concat("cookie:", web.AuthHeaderName),
		}
		wapi.Use(echojwt.WithConfig(config))
	}

	wapi.GET("/config", s.factory.WebHandler.Config.Load)
	wapi.POST("/config", s.factory.WebHandler.Config.Save)

	wapi.GET("/ws", s.factory.WebHandler.WS.Register)

	// Device API
	mapi := s.factory.Webs.Group("/api/device")
	mapi.Use(echojwt.JWT([]byte(s.factory.Config.API.TokenSecretKey)))

	mapi.GET("/ws", s.factory.APIHandler.WS.Register)
}
