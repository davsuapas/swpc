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

	"github.com/labstack/echo/v4"
	"github.com/swpoolcontroller/internal/micro"
)

// Stream exchanges information with the micro controller
type Stream struct {
	control *micro.Controller
}

func NewStream(control *micro.Controller) *Stream {
	return &Stream{
		control: control,
	}
}

// Status gets information on how the micro controller should behave
func (s *Stream) Status(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, s.control.Status())
}

// Download transfers the metrics between micro controller and the web
func (s *Stream) Download(ctx echo.Context) error {
	s.control.Download(ctx.FormValue("metrics"))

	return ctx.JSON(http.StatusOK, s.control.Status())
}
