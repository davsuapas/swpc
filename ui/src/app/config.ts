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

export interface AppConfig {
    authLoginUrl: string
    authLogoutUrl: string
    checkAuthName: string
    iotConfig: boolean
    aiSample: boolean
}

const keyAppConfig = "app-config";

// loadAppConfig loads the server app config to session cache
export async function loadAppConfig(): Promise<Boolean> {
    
    try {
        const res = await fetch("/app/config", {method: "GET"});
        if (res.status == 200) {
            sessionStorage.setItem(keyAppConfig, await res.text());

            return true;
        }
    } catch (ex) {
        console.log("app.loadConfig. Loading server configuration: " + ex);
    }

    return false;
}

// config gets server app config from session cache
export function appConfig(): AppConfig {
    const res = sessionStorage.getItem(keyAppConfig);

    if (res) {
        return JSON.parse(res);
    }

    console.log("appConfig. The app configuration is empty");

    return {
      authLoginUrl: "",
      authLogoutUrl: "",
      checkAuthName: "",
      iotConfig: true,
      aiSample: true
    };
}