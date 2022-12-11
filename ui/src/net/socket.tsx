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
import { Websocket, WebsocketBuilder } from "websocket-ts/lib";
import User from "../login/user";
import Alert from "../support/alert";

enum CommStatus {
    inactiveComm,
    breakComm
}

export default class SocketFactory {

    private user: User

    private ws: WebsocketBuilder;

    constructor(private alert: RefObject<Alert>, private wait: React.Dispatch<React.SetStateAction<boolean>>) {
        const protocol = location.protocol == "https:" ? "wss" : "ws";
        this.ws = new WebsocketBuilder(protocol + "://" + document.location.host + "/web/api/ws");
        this.user = new User()
    }

    // start opens socket connection, registers in server and controls events
    open(): Websocket {
        return this.ws.onClose((_, ev) => {
            console.log("El socket se ha cerrado con el código: " + ev.code);

            if (this.alert.current) {
                this.alert.current.content(
                    "Conexión cerrada",
                    "Puede ser que haya caducado la sesión o se haya producido un problema de comunicación," +
                    "inténtelo más tarde. " +
                    "Se procederá a cerrar la sessión de trabajo.");
                this.alert.current.events.closed = () => {
                    this.user.logoff();
                };
                this.alert.current.open();
            }
        })
        .onError((_ , ev) => {
            if (this.alert.current) {
                this.alert.current.content(
                    "Error de conexión",
                    "Se ha producido un error con la conexión en tiempo real. " +
                    "Se procederá a cerrar la sessión de trabajo.");
                this.alert.current.events.closed = () => {
                    this.user.logoff()
                };
                this.alert.current.open();
            }
        })
        .onMessage((socket , ev) => {
            const message = new MessageFactory(ev.data)

            if (message.messageType == MessageType.control ) {
                const status = message.controlMessage();

                if (status == CommStatus.inactiveComm) {
                    if (this.alert.current) {
                        this.alert.current.content(
                            "Comunicación con el micro controlador sin respuesta",
                            "No se detecta ningún envío de métricas del micro controlador, seguiremos " +
                            "intentado reestablecer la comunicación. No cierre ninguna vetana mientras " + 
                            "se restablece la comunicación");
                        this.alert.current.open();
                        this.wait(true)
                    }
                } else {
                    if (this.alert.current) {
                        this.alert.current.content(
                            "Parece que la comunicación con el micro controlador se encuentra caída",
                            "Demsasiado tiempo sin respuesta, asegúrese que el micro controlador " +
                            "se encuentra encedido y que la comunicación se encuentra habilitada. " + 
                            "También puede ser debido, a que no se encuentra dentro del horario establecido " +
                            "para la recepción de las métricas o simplemente hay un retraso en las comunicaciones. " +
                            "En este caso último, inténtelo más tarde. " +
                            "Se procederá a cerrar la sessión de trabajo.");
                        this.alert.current.events.closed = () => {
                            socket.close();
                            this.user.logoff();
                        };
                        this.alert.current.open();
                    }
                }
            }
        }).build();
    }
}

enum MessageType {
    control,
    meassure  
}

class MessageFactory {
    messageType: MessageType;
    private rawMessage: string;

    constructor(msg: string) {
        const tokens = msg.split(":")
        this.messageType = tokens[0] == "0" ? MessageType.control : MessageType.meassure
        this.rawMessage = tokens[1]
    }

    controlMessage(): CommStatus {
        return Number.parseInt(this.rawMessage) == 1 ? CommStatus.inactiveComm : CommStatus.breakComm
    }
}