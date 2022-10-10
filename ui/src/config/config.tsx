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
import green from '@mui/material/colors/green';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Slide from '@mui/material/Slide';
import TextField from '@mui/material/TextField';
import { TransitionProps } from '@mui/material/transitions';
import React, { forwardRef, useRef } from 'react';
import Fetch from '../net/fetch';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const defaultCheckSend = 1;
const defaultimeIniSend = "11:00";
const defaultimeEndSend = "12:00";
const defaultWakeUp = 20;
const defaultBuffer = 10;

// Config displays a configuration form
export default forwardRef( (props: any, ref: any) => {
    const fetch = new Fetch(props.alert);

    const [openv, setOpenv] = React.useState(false);

    const [checkSendValid, setCheckSendValid] = React.useState(true);
    const [checkSendValue, setCheckSendValue] = React.useState(defaultCheckSend);

    const [timeSendValid, setTimeSendValid] = React.useState(true);

    const [timeIniSendValue, setTimeIniSendValue] = React.useState(defaultimeIniSend);
    const [timeEndSendValue, setTimeEndSendValue] = React.useState(defaultimeEndSend);

    const [wakeUpValid, setWakeUpValid] = React.useState(true);
    const [wakeUpValue, setWakeUpValue] = React.useState(defaultWakeUp);

    const [bufferValid, setBufferValid] = React.useState(true);
    const [bufferValue, setBufferValue] = React.useState(defaultBuffer);

    const [saving, setSaving] = React.useState(false);

    // init initializes the form when is invoked
    function init() {
        setSaving(false);
        initControl();
    }

    function initControl() {
        setCheckSendValid(true);
        setTimeSendValid(true);
        setWakeUpValid(true);
        setBufferValid(true);
    }

    function valid(): boolean {
        initControl();

        let valid = true;

        if (!(checkSendValue >= 1 && checkSendValue <= 10)) {
            setCheckSendValid(false);
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

        if (!(wakeUpValue >= 5 && wakeUpValue <= 120)) {
            setWakeUpValid(false);
            valid = false;
        }

        if (!(bufferValue >= 3 && bufferValue <= 60)) {
            setBufferValid(false);
            valid = false;
        }

        return valid;
    }

    // setControl sets the data configuration into input
    function setControl(checkSend: number, timeIniSend: string, timeEndSend: string, setWakeUp: number, buffer: number) {
        setCheckSendValue(checkSend);
        setTimeIniSendValue(timeIniSend);
        setTimeEndSendValue(timeEndSend);
        setWakeUpValue(setWakeUp);
        setBufferValue(buffer);
    }

    // open opens the configuration form with the datas send from server
    const open = async (setloadingConfig: any) => {
      init();

      fetch.send("/web/api/config", {
        method: "GET"
      },
      async (result: Response) => {
        setloadingConfig(false);
        if (result.ok) {
            try {
                const res = await result.json();
                setControl(res.checkSend, res.timeIniSend, res.timeEndSend, res.wakeUp, res.buffer);
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
        if (result.status == 404) {
            setControl(defaultCheckSend, defaultimeIniSend, defaultimeEndSend, defaultWakeUp, defaultBuffer);
            setOpenv(true);
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
          body: JSON.stringify({
            "checkSend": checkSendValue,
            "timeIniSend": timeIniSendValue,
            "timeEndSend": timeEndSendValue,
            "wakeUp": wakeUpValue,
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
                    El micro conrolador encargado de medir los valores de la piscina, puede consumir bastante batería.
                    Para salvaguardar la bateria, es conveniente configurar las horas de emisión de las métricas
                    por parte del micro. En esta sección se le pedirá dos grupos de configuración.
                    El primero consiste en configurar, cada cuantas horas el micro chequea entre que horas emitirá
                    métricas y en el segundo se configurará cuatro valores, entre que horas se emitirán los valores
                    de las métricas y en caso de que no exista ningún receptor para recibir métricas, cada cuantos segundos
                    se despierta el micro para comprobar cuando hay un receptor o receptores disponibles,
                    siempre teniendo en cuenta las horas establecidas. También se establece cuantos segundos almacena
                    el micro las métricas antes de enviarlas. Cuanto más tiempo tiene la web abierta
                    para recibir métricas mayor debería ser este buffer.
                </DialogContentText>
                <TextField
                    id="checkSend"
                    sx={{marginTop:"30px"}}
                    label="Chequeo emisión"
                    type="number"
                    value={checkSendValue}
                    onChange={event => setCheckSendValue(parseInt(event.target.value))}
                    InputProps={{ inputProps: { min: 1, max: 10 } }}
                    size="small"
                    error={!checkSendValid}
                    helperText={checkSendValid ?
                         "Cada cuantas horas chequa entre que horas emite" :
                         "El valor debe estar entre 1 y 10 minutos"}
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
                    <TextField
                        id="wakeUp"
                        sx={{marginTop:"20px"}}
                        label="Chequeo receptores"
                        type="number"
                        value={wakeUpValue}
                        onChange={event => setWakeUpValue(parseInt(event.target.value))}
                        InputProps={{ inputProps: { min: 5, max: 120 } }}
                        size="small"
                        error={!wakeUpValid}
                        helperText={wakeUpValid ?
                            "Cada cuantas segundos se despierta el micro" :
                            "El valor debe estar entre 5 y 120 segundos"}
                    />
                </FormGroup>                                        
                <FormGroup row sx={{marginTop:"20px"}}>
                    <TextField
                        id="buffer"
                        label="Buffer"
                        type="number"
                        value={bufferValue}
                        onChange={event => setBufferValue(parseInt(event.target.value))}
                        InputProps={{ inputProps: { min: 3, max: 60 } }}
                        size="small"
                        error={!bufferValid}
                        helperText={bufferValid ?
                            "Tiempo que almacena el micro las métricas antes de enviar" :
                            "El valor debe estar entre 3 y 60 segundos"}
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
                                color: green[900],
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
 