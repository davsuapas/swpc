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


// Config displays a configuration form
export default forwardRef( (props: any, ref: any) => {
    const fetch = new Fetch(props.alert);

    const [openv, setOpenv] = React.useState(false);

    const [wakeupValid, setWakeupValid] = React.useState(true);
    const [wakeupValue, setWakeupValue] = React.useState(0);

    const [timeSendValid, setTimeSendValid] = React.useState(true);

    const [timeIniSendValue, setTimeIniSendValue] = React.useState("");
    const [timeEndSendValue, setTimeEndSendValue] = React.useState("");

    const [bufferValid, setBufferValid] = React.useState(true);
    const [bufferValue, setBufferValue] = React.useState(0);

    const [saving, setSaving] = React.useState(false);

    // init initializes the form when is invoked
    function init() {
        setSaving(false);
        initControl();
    }

    function initControl() {
        setWakeupValid(true);
        setTimeSendValid(true);
        setBufferValid(true);
    }

    function valid(): boolean {
        initControl();

        let valid = true;

        if (!(wakeupValue >= 1 && wakeupValue <= 10)) {
            setWakeupValid(false);
            valid = false;
        }

        const timeIni = parseTime(timeIniSendValue);
        const timeEnd = parseTime(timeEndSendValue);

        if (!timeIni || !timeEnd) {
            setTimeSendValid(false);
            valid = false;
        } else {
            if (timeEnd < timeIni) {
                setTimeSendValid(false);
                valid = false;
            }
        }

        if (!(bufferValue >= 3 && bufferValue <= 60)) {
            setBufferValid(false);
            valid = false;
        }

        return valid;
    }

    // setControl sets the data configuration into input
    function setControl(wakeup: number, timeIniSend: string, timeEndSend: string, buffer: number) {
        setWakeupValue(wakeup);
        setTimeIniSendValue(timeIniSend);
        setTimeEndSendValue(timeEndSend);
        setBufferValue(buffer);
    }

    // open opens the configuration form with the datas send from server
    const open = async (setloadingConfig: any) => {
      setloadingConfig(true);
      init();

      fetch.send("/web/api/config", {
        method: "GET"
      },
      async (result: Response) => {
        setloadingConfig(false);
        if (result.ok) {
            try {
                const res = await result.json();
                setControl(res.wakeup, res.timeIniSend, res.timeEndSend, res.buffer);
                setOpenv(true);
            }
            catch {
                props.alert.current.content(
                    "Error interno",
                    "Se ha producido un error al recuperar la configuración. Vuelva a intentarlo más tarde");
                    props.alert.current.open();
            }
            return true;
        }
        return false;
      },
      () => {
        setloadingConfig(false);
      });
    };
  
    // close saves the configuration in the server an close the window
    async function close(save: boolean) {
      if (!save) {
        setOpenv(false);
        return
      } 

      if (valid()) {
        setSaving(true);

        fetch.send("/web/api/config", {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "wakeup": wakeupValue,
            "iniSendTime": timeIniSendValue,
            "endSendTime": timeEndSendValue,
            "buffer": bufferValue
          })
        },
        async (result: Response) => {
          setOpenv(false);
          if (result.ok) {
            return true;
          }
          return false;
        },
        () => {
            setOpenv(false);
        });
      }
    }
    
    // Export the function
    if (ref) ref.current = {open}

    return (
    <div>
        <Dialog
            open={openv}
            TransitionComponent={Transition}
            keepMounted
            onClose={() => !saving && close(false)}
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
                    value={wakeupValue}
                    onChange={event => setWakeupValue(parseInt(event.target.value))}
                    InputProps={{ inputProps: { min: 15, max: 120 } }}
                    size="small"
                    error={!wakeupValid}
                    helperText={wakeupValid ?
                         "Cada cuantos minutos chequa entre que horas emite" :
                         "El valor debe estar entre 15 y 120 minutos"}
                />                
                <FormGroup row sx={{marginTop:"30px"}}>
                    <TextField
                        id="timeIniSend"
                        sx={{marginTop:"10px"}}
                        label="Hora inicio"
                        type="time"
                        value={timeIniSendValue}
                        onChange={event => setTimeIniSendValue(event.target.value)}
                        size="small"
                        error={!timeSendValid}
                        helperText={timeSendValid ?
                            "Hora de inicio para el envío de métricas" :
                            "El valor debe ser una hora y debe ser menor la hora inicio de la de fín"}
                    />                
                    <TextField
                        id="timeEndSend"
                        sx={{marginTop:"10px"}}
                        label="Hora fin"
                        type="time"
                        value={timeEndSendValue}
                        onChange={event => setTimeEndSendValue(event.target.value)}
                        size="small"
                        error={!timeSendValid}
                        helperText={timeSendValid ?
                            "Hora que finaliza la emisión de las métricas" :
                            "El valor debe ser una hora y debe ser menor la hora inicio de la de fín"}
                    />                
                </FormGroup>                                        
                <FormGroup row sx={{marginTop:"20px"}}>
                    <TextField
                        id="buffer"
                        label="Buffer"
                        type="number"
                        value={bufferValue}
                        onChange={event => setBufferValue(parseInt(event.target.value))}
                        InputProps={{ inputProps: { min: 3, max: 20 } }}
                        size="small"
                        error={!bufferValid}
                        helperText={bufferValid ?
                            "Tiempo que almacena el micro las métricas antes de enviar (buffer)" :
                            "El valor debe estar entre 3 y 20 segundos"}
                    />                
                </FormGroup>                    
            </DialogContent>
            <DialogActions>
                <Box sx={{ m: 1, position: 'relative' }}>
                    <Button onClick={() => close(true)}>{saving ? "Guardando" : "Guardar"}</Button>
                    {saving && (
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
                <Button disabled={saving} onClick={() => close(false)}>Cancelar</Button>
            </DialogActions>
        </Dialog>
    </div>
    );
});

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
 