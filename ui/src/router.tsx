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

import { AppConfig, appConfig } from "./app/config";
import Meter from "./info/meter";

// PrivateRoute creates un protected component
export default function PrivateRoute({ children }: { children: JSX.Element } ) {
  const config = appConfig()
  
  if (isAuthenticated(config)) {
    return children;
  }

  window.location.href = config.authLoginUrl;

  return (<Meter message="Redirigiendo a la paǵina de login..."/>);
}

function isAuthenticated(config: AppConfig)  {
  return getCookie(config.checkAuthName);
}

function getCookie(cname: string) {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');

  for(let i = 0; i <ca.length; i++) {
    let c = ca[i];

    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }

    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }

  return null;
}