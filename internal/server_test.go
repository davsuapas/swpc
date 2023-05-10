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

package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swpoolcontroller/internal"
	"github.com/swpoolcontroller/internal/config"
)

func TestServer_Start(t *testing.T) {
	t.Parallel()

	f := internal.NewFactory()
	s := internal.NewServer(f)

	go func() {
		s.Kill()
	}()

	assert.NoError(t, s.Start())
}

func TestServer_Middleware(t *testing.T) {
	t.Parallel()

	f := internal.NewFactory()
	s := internal.NewServer(f)

	s.Middleware()

	assert.NotNil(t, f.Webs)
}

func TestServer_Route_Auth_Provider_Oauth2(t *testing.T) {
	t.Parallel()

	f := internal.NewFactory()
	f.Config.Auth.Provider = config.AuthProviderOauth2
	s := internal.NewServer(f)

	s.Route()

	assert.Equal(t, len(f.Webs.Router().Routes()), 53)
}

func TestServer_Route_Auth_Provider_Dev(t *testing.T) {
	t.Parallel()

	f := internal.NewFactory()
	s := internal.NewServer(f)

	s.Route()

	assert.Equal(t, len(f.Webs.Router().Routes()), 31)
}
