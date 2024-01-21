/*
 *   Copyright (c) 2022 ELIPCERO
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

import { Root } from 'react-dom/client';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import Dashboard from '../dashboard/dashboard';
import AuthError, { urlAuthError } from '../auth/auth_error';
import Logout from '../auth/logout';
import PrivateRoute from '../router'
import { loadAppConfig } from './config';
import Ups from '../info/ups';

export default async function main(root: Root) {

    const loaded = await loadAppConfig()

    if (loaded) {
      root.render(
          <BrowserRouter>
              <Routes>
                  <Route path={urlAuthError} element={
                    <AuthError />
                  }/>
                  <Route path="/auth/logout" element={
                    <Logout redirectURL="/"/>
                  }/>
                  <Route path="/" element={
                    <PrivateRoute>
                      <Dashboard />
                    </PrivateRoute>
                  }/>
              </Routes>
          </BrowserRouter>,
      );
    } else {
      root.render(
        <Ups message="Se ha producido un problema cargando la configuración del servidor"/>
      );
    }
}