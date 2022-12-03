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

import { RefObject } from "react";
import User from "../login/user";
import Alert from "../support/alert";

// Fetch communicates using fetch
export default class Fetch {

    private user: User

    constructor(private alert: RefObject<Alert>) {
        this.user = new User()
    }

    // send sends http request and allows to handle the response through a function.
    // If not handled by the callback it displays alerts on screen
    async send(url: string, props: {}, actions: (res: Response) => Promise<boolean>, error: any = null) {
        try {
            const res = await fetch(url, props);
            const manejado = await actions(res);

            if (!manejado) {
                switch (res.status) {
                    case 400:
                        this.alert.current?.content(
                            "Error interno",
                            "Se ha producido un error se solicitud errónea. Vuelva a intentarlo más tarde.");
                        this.alert.current?.open();
                        break;
                    case 401:
                        if (this.alert.current) {    
                            this.alert.current.content(
                                "Error de seguridad",
                                "La sessión ha caducado. Se procederá a cerrar la sessión de trabajo.");
                            this.alert.current.events.closed = () => {
                                this.user.logoff()
                            };
                            this.alert.current.open();
                        }
                        break;
                    case 500:
                        this.alert.current?.content(
                            "Error interno",
                            "Se ha producido un error interno en el servidor. Vuelva a intentarlo más tarde.");
                        this.alert.current?.open();
                        break;
                    default:
                        this.alert.current?.content(
                            "Error no controlado",
                            "Se ha producido un error no controlado. Consulte con el proveedor del servicio.");
                        this.alert.current?.open();
                        break;
                }
            }
        } catch (ex) {
            console.log("Fetch. Web request error: " + ex);

            this.alert.current?.content(
                "Error inesperado",
                "Se ha producido un error inesperado al comunicar con el servidor. Vuelva a intentarlo más tarde.");
            this.alert.current?.open();
            
            if (typeof error === 'function') {
                error();
            }
        }
    }
}