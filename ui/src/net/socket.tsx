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

import { RefObject } from "react";
import { Websocket, WebsocketBuilder } from "websocket-ts/lib";
import { Actions } from "../dashboard/dashboard";
import { logoff } from "../auth/user";
import Alert from "../info/alert";

// CommStatus is the communications status
export enum CommStatus {
    inactive,
    active,
    broadcasting
}

export interface Metrics {
    temp: number[];
    ph: number[];
    orp: number[];
}

export interface SocketEvent {
    streamMetrics: (metrics: Metrics) => void;
  }
 
// SocketFactory Manages socket iteration with the server
export default class SocketFactory {

    event: SocketEvent;

    private ws: WebsocketBuilder;

    state: CommStatus;

    constructor(
      private alert: RefObject<Alert>,
      private actions: Actions) {

      this.state = CommStatus.inactive;

      const protocol = location.protocol == "https:" ? "wss" : "ws";
      this.ws = new WebsocketBuilder(protocol + "://" + document.location.host + "/api/web/ws");

      this.event = {
          streamMetrics: () => {}
      }
    }

    // start opens socket connection, registers in server and controls events
    open(): Websocket {
        return this.ws.onClose((_, ev) => {
            console.log("El socket se ha cerrado con el código: " + ev.code);

            this.state = CommStatus.inactive;

            if (this.alert.current) {
                this.alert.current.content(
                    "Conexión cerrada",
                    "La sesión se ha caducado, o bien por cumplir el tiempo máximo de sesión, " +
                    "o bien porque el servidor, por algún motivo, ha cerrado la sesión. " + 
                    "Si desea continuar vuelva a iniciar sesión. " +
                    "Se procederá a cerrar la sessión de trabajo.");
                this.alert.current.events.closed = () => {
                    logoff();
                };
                this.alert.current.open();
            }
        })
        .onError((_ , ev) => {
            this.state = CommStatus.inactive;

            if (this.alert.current) {
                this.alert.current.content(
                    "Error de conexión",
                    "Se ha producido un error con la conexión en tiempo real. " +
                    "Se procederá a cerrar la sessión de trabajo.");
                this.alert.current.events.closed = () => {
                    logoff()
                };
                this.alert.current.open();
            }
        })
        .onMessage((socket , ev) => {
            const message = new MessageFactory(ev.data)

            try {
                if (message.messageType == MessageType.control ) {
                    this.state = message.controlMessage();

                    if (this.alert.current) {
                        this.alert.current.content(
                            "Comunicación con el micro-controlador sin respuesta",
                            "No se detecta ningún envío de métricas desde el micro-controlador, seguiremos " +
                            "intentado reestablecer la comunicación. Si persiste el problema, " +
                            "asegúrese que el micro-controlador se encuentra encedido y que la comunicación " +
                            "se encuentra habilitada. También puede ser debido, a que no se encuentra " +
                            "dentro del horario establecido para la recepción de las métricas, " +
                            "o simplemente hay un retraso en las comunicaciones")
                            
                        this.alert.current.open();
                        this.actions.activeStandby(true)
                    }
                } else {
                    this.state = CommStatus.broadcasting;
                    this.actions.activeStandby(false);
                    this.event.streamMetrics(message.metricsMessage());
                }
            }
            catch (ex) {
                console.log("Sockets.onMessage: " + ex);

                this.state = CommStatus.inactive;

                this.alert.current?.content(
                    "Se ha producido un error al recibir información del servidor.",
                    "Si el error persiste, cierre la sesión y vuelva a intentarlo")

                this.alert.current?.open();
            }
        }).build();
    }
}

enum MessageType {
    // control is a message of control type
    control,
    // control is a message of metric type
    metrics  
}

// MessageFactory builds the message
class MessageFactory {
    messageType: MessageType;
    private rawMessage: string;

    constructor(msg: string) {
        this.messageType = msg.at(0) == "0" ? MessageType.control : MessageType.metrics
        this.rawMessage = msg.substring(1)
    }

    // controlMessage gets communication status
    controlMessage(): CommStatus {
        return Number.parseInt(this.rawMessage) == 1 ?
         CommStatus.active : 
         CommStatus.inactive
    }

    metricsMessage(): Metrics {
        const metrics = JSON.parse(this.rawMessage);
        return {
            temp: metrics.temp.map((m: String) => Number(m)),
            ph: metrics.ph.map((m: String) => Number(m)),
            orp: metrics.orp.map((m: String) => Number(m)),
        }
    }
}