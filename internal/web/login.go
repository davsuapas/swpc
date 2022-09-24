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
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Login controllers the access of the user
type Login struct {
	log *zap.Logger
}

// Submit validates the user, generates token and save session
func (l *Login) Submit(c echo.Context) error {
	email := c.FormValue("email")
	pass := c.FormValue("password")
}

// Users controllers the user of app in memory
type users struct {
	user map[string]string
}

func NewUsersInMemory() users {
	return users{
		user: map[string]string{
			"dav.sua.pas@gmail.com": "RCrkRDBG6cc=",
		},
	}
}
