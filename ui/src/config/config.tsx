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

import { Box, CircularProgress, FormGroup } from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Slide from '@mui/material/Slide';
import TextField from '@mui/material/TextField';
import { TransitionProps } from '@mui/material/transitions';
import React, { forwardRef } from 'react';
import { Actions } from '../dashboard/dashboard';
import Fetch from '../net/fetch';
import { colorPurple } from '../support/color';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="up" ref={ref} {...props} />;
});

interface ConfigState {
    open: boolean;
    wakeupValid: boolean;
    wakeupValue: number;
    sendTimeValid: boolean;
    iniSendTimeValue: string;
    endSendTimeValue: string;
    bufferValid: boolean;
    bufferValue: number;
    saving: boolean;
}

export default class Config extends React.Component<any, ConfigState> {

    private fetch?: Fetch;

    constructor(props: any) {
        super(props);

        this.fetch = new Fetch(this.props.alert);

        this.state = {
            open: false,
            wakeupValid: false,
            wakeupValue: 0,
            sendTimeValid: false,
            iniSendTimeValue: "",
            endSendTimeValue: "",
            bufferValid: false,
            bufferValue: 0,
            saving: false
        };
    }

    // open opens the configuration form with the datas send from server
    async open(actions: Actions) {
        actions.activeLoadingConfig(true);
        this.init();
    
        this.fetch?.send("/web/api/config", {
            method: "GET"
        },
        async (result: Response) => {
            actions.activeLoadingConfig(false);
            if (result.ok) {
                try {
                    const res = await result.json();
                    this.setControl(res.wakeup, res.iniSendTime, res.endSendTime, res.buffer);
                    this.setState({open: true});
                }
                catch {
                    this.props.alert.current.content(
                        "Error interno",
                        "Se ha producido un error al recuperar la configuración. Vuelva a intentarlo más tarde");
                        this.props.alert.current.open();
                }
                return true;
            }
            return false;
        },
        () => {
            actions.activeLoadingConfig(false);
        });
    };
      
    // init initializes the form when is invoked
    private init() {
        this.setState({saving: false});
        this.initControl();
    }

    private initControl() {
        this.setState({
            wakeupValid: true,
            sendTimeValid: true,
            bufferValid: true,
        });
    }

    private valid(): boolean {
        this.initControl();

        let valid = true;

        if (!(this.state.wakeupValue >= 15 && this.state.wakeupValue <= 120)) {
            this.setState({wakeupValid: false});
            valid = false;
        }

        const timeIni = parseTime(this.state.iniSendTimeValue);
        const timeEnd = parseTime(this.state.endSendTimeValue);

        if (!timeIni || !timeEnd) {
            this.setState({sendTimeValid: false});
            valid = false;
        } else {
            if (timeEnd < timeIni) {
                this.setState({sendTimeValid: false});
                valid = false;
            }
        }

        if (!(this.state.bufferValue >= 3 && this.state.bufferValue <= 60)) {
            this.setState({bufferValid: false});
            valid = false;
        }

        return valid;
    }

    // setControl sets the data configuration into input
    private setControl(wakeup: number, iniSendTime: string, endSendTime: string, buffer: number) {
        this.setState({
            wakeupValue: wakeup,
            iniSendTimeValue: iniSendTime,
            endSendTimeValue: endSendTime,
            bufferValue: buffer}
        );
    }
  
    // close saves the configuration in the server an close the window
    private async close(save: boolean) {
      if (!save) {
        this.setState({open: false});
        return;
      } 

      if (this.valid()) {
        this.setState({saving: true});

        this.fetch?.send("/web/api/config", {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "wakeup": this.state.wakeupValue,
            "iniSendTime": this.state.iniSendTimeValue,
            "endSendTime": this.state.endSendTimeValue,
            "buffer": this.state.bufferValue
          })
        },
        async (result: Response) => {
            this.setState({open: false});
            if (result.ok) {
                return true;
            }
            return false;
        },
        () => {
            this.setState({open: false});
        });
      }
    }

    render(): React.ReactNode {
        return (
        <div>
            <Dialog
                open={this.state.open}
                TransitionComponent={Transition}
                keepMounted
                onClose={() => !this.state.saving && this.close(false)}
            >
                <DialogTitle>Configuración</DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-slide-description">
                        El micro controlador encargado de medir los valores de la piscina, puede consumir bastante batería.
                        Para salvaguardar la bateria, es conveniente configurar las horas de emisión de las métricas
                        por parte del micro. En esta sección se le pedirá dos grupos de configuración.
                        El primero consiste en configurar, cada cuantos minutos el micro chequea entre que horas emitirá
                        métricas y en el segundo se configurará tres valores, entre que horas se emitirán los valores
                        de las métricas y cada cuantos segundos almacena el micro las métricas antes de enviarlas. 
                        Cuanto más tiempo tiene la web abierta para recibir métricas mayor debería ser este buffer.
                    </DialogContentText>
                    <TextField
                        id="wakeup"
                        sx={{marginTop:"30px"}}
                        label="Chequeo emisión"
                        type="number"
                        value={this.state.wakeupValue}
                        onChange={event => this.setState({wakeupValue: parseInt(event.target.value)})}
                        InputProps={{ inputProps: { min: 15, max: 120 } }}
                        size="small"
                        error={!this.state.wakeupValid}
                        helperText={this.state.wakeupValid ?
                            "Cada cuantos minutos chequa entre que horas emite" :
                            "El valor debe estar entre 15 y 120 minutos"}
                    />                
                    <FormGroup row sx={{marginTop:"30px"}}>
                        <TextField
                            id="iniSendTime"
                            sx={{marginTop:"10px"}}
                            label="Hora inicio"
                            type="time"
                            value={this.state.iniSendTimeValue}
                            onChange={event => this.setState({iniSendTimeValue: event.target.value})}
                            size="small"
                            error={!this.state.sendTimeValid}
                            helperText={this.state.sendTimeValid ?
                                "Hora de inicio para el envío de métricas" :
                                "El valor debe ser una hora y debe ser menor la hora inicio de la de fín"}
                        />                
                        <TextField
                            id="endSendTime"
                            sx={{marginTop:"10px"}}
                            label="Hora fin"
                            type="time"
                            value={this.state.endSendTimeValue}
                            onChange={event => this.setState({endSendTimeValue: event.target.value})}
                            size="small"
                            error={!this.state.sendTimeValid}
                            helperText={this.state.sendTimeValid ?
                                "Hora que finaliza la emisión de las métricas" :
                                "El valor debe ser una hora y debe ser menor la hora inicio de la de fín"}
                        />                
                    </FormGroup>                                        
                    <FormGroup row sx={{marginTop:"20px"}}>
                        <TextField
                            id="buffer"
                            label="Buffer"
                            type="number"
                            value={this.state.bufferValue}
                            onChange={event => this.setState({bufferValue: parseInt(event.target.value)})}
                            InputProps={{ inputProps: { min: 3, max: 20 } }}
                            size="small"
                            error={!this.state.bufferValid}
                            helperText={this.state.bufferValid ?
                                "Tiempo que almacena el micro las métricas antes de enviar (buffer)" :
                                "El valor debe estar entre 3 y 20 segundos"}
                        />                
                    </FormGroup>                    
                </DialogContent>
                <DialogActions>
                    <Box sx={{ m: 1, position: 'relative' }}>
                        <Button onClick={() => this.close(true)}>{this.state.saving ? "Guardando" : "Guardar"}</Button>
                        {this.state.saving && (
                            <CircularProgress
                                size={24}
                                sx={{
                                    color: colorPurple,
                                    position: 'absolute',
                                    top: '50%',
                                    left: '50%',
                                    marginTop: '-12px',
                                    marginLeft: '-12px',
                                }}
                            />
                        )}                
                    </Box>
                    <Button disabled={this.state.saving} onClick={() => this.close(false)}>Cancelar</Button>
                </DialogActions>
            </Dialog>
        </div>
        );
    }
}

function parseTime(t: string): Date | null {
    const d = new Date();
    const time = t.match(/(\d+)(?::(\d\d))?\s*(p?)/);
    if (!time) {
        return null;
    }
    d.setHours(parseInt(time[1]) + (time[3] ? 12 : 0));
    d.setMinutes(parseInt(time[2]) || 0);
    return d;
 }
 