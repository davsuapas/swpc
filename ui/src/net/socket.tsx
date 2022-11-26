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
import { useNavigate } from "react-router-dom";
import { Websocket, WebsocketBuilder } from "websocket-ts/lib";
import Alert from "../support/alert";

export default class SocketFactory {

    private navigate = useNavigate();

    private ws: WebsocketBuilder

    constructor(private alert: RefObject<Alert>) {
        const protocol = location.protocol == "https:" ? "wss" : "ws";
        this.ws = new WebsocketBuilder(protocol + "://" + document.location.host + "/web/api/ws");
    }

    // start opens socket connection, registers in server and controls events
    open(): Websocket {
        return this.ws.onClose((_, ev) => {
            console.log("El socket se ha cerrado con el código: " + ev.code + ", motivo: " + ev.reason);

            this.alert.current?.content(
                "Conexión cerrada",
                "Puede ser que haya caducado la sesión o se haya producido un problema de comunicación," +
                "inténtelo más tarde. " +
                "Se procederá a cerrar la sessión de trabajo.");
            if (this.alert.current) {
                this.alert.current.events.closed = () => {
                    this.navigate("/");
                };
            }
            this.alert.current?.open();
        })
        .onError((_ , ev) => {
            this.alert.current?.content(
                "Error de conexión",
                "Se ha producido un error con la conexión en tiempo real. " +
                "Se procederá a cerrar la sessión de trabajo.");
            if (this.alert.current) {
                this.alert.current.events.closed = () => {
                    this.navigate("/");
                };
            }
            this.alert.current?.open();
        }).build();
    }
}