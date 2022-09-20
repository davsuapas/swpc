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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	factory *Factory
}

func NewServer(factory *Factory) *Server {
	return &Server{
		factory: factory,
	}
}

// Start starts the graceful http server
func (s *Server) Start() {
	// Start server
	go func() {
		if err := s.factory.Webs.Start(s.factory.Config.Address()); err != nil {
			s.factory.Logger.Info("Shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.factory.Webs.Shutdown(ctx); err != nil {
		s.factory.Logger.Error(err.Error())
	}
}

// Middleware configure security and behaviour of http
func (s *Server) Middleware() {
	s.factory.Webs.Use(middleware.Recover())

	// Logger
	s.factory.Webs.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := []zapcore.Field{
				zap.String("request_id", v.RequestID),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("host", v.Host),
				zap.String("Latency", v.Latency.String()),
				zap.String("method", v.Method),
				zap.String("request_uri", v.Method),
				zap.Int("status", v.Status),
				zap.Int64("response_size", v.ResponseSize),
				zap.String("user_agent", v.UserAgent),
			}

			switch {
			case v.Status >= 500:
				s.factory.Logger.Error("Web server error", fields...)
			case v.Status >= 400:
				s.factory.Logger.Error("Web client error", fields...)
			case v.Status >= 300:
				s.factory.Logger.Info("Web server redirection", fields...)
			default:
				s.factory.Logger.Info("Web success server response", fields...)
			}
			return nil
		},
	}))
}
