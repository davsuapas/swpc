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

import React from 'react';
import { TransitionProps } from '@mui/material/transitions';
import Slide from '@mui/material/Slide';
import Fetch from '../net/fetch';
import { Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, FormGroup, Grid, InputLabel, MenuItem, Paper, Select, styled } from '@mui/material';
import { Stack, TextField } from '@mui/material';
import { DialogTitle } from '@mui/material';
import Meassure from '../dashboard/meassure';
import * as literals from '../support/literals';
import { colorPurple } from '../support/color';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const Item = styled(Paper)(({ theme }) => ({
  backgroundColor: theme.palette.mode === 'dark' ? '#1A2027' : '#fff',
  ...theme.typography.body2,
  padding: theme.spacing(1),
  textAlign: 'center',
  color: theme.palette.text.secondary,
}));

interface SampleState {
  open: boolean;
  cl: number;
  clValid: boolean;
  waterQuality: number;
  saving: boolean;
}

export default class Sample extends React.Component<any, SampleState> {

  private fetch?: Fetch;

  private meassureTemp: React.RefObject<Meassure>;
  private meassurePh: React.RefObject<Meassure>;
  private meassureOrp: React.RefObject<Meassure>;

  constructor(props: any) {
      super(props);

      this.fetch = new Fetch(this.props.alert);

      this.state = {
        open: false,
        cl: 0,
        clValid: true,
        waterQuality: 0,
        saving: false
      };

      this.meassureTemp = React.createRef<Meassure>();
      this.meassurePh = React.createRef<Meassure>();
      this.meassureOrp = React.createRef<Meassure>();
  }

  // open opens the samples editor with the values sent by micro
  open(temp: number, ph: number, orp: number) {
    this.setState({
      cl: 0,
      clValid: true,
      waterQuality: 0,
      saving: false
    });

    this.meassureTemp.current?.setMeassure(temp);
    this.meassurePh.current?.setMeassure(ph);
    this.meassureOrp.current?.setMeassure(orp);

    this.setState({open: true});
  }

  // close saves the sample in the server an close the window
  private async close(save: boolean) {
    if (!save) {
      this.setState({open: false});
      return;
    } 

    if (this.valid()) {
      this.setState({saving: true});

      this.fetch?.send("/api/web/sample", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          "temp": this.meassureTemp.current?.state.value,
          "ph": this.meassurePh.current?.state.value,
          "orp": Number(this.meassureOrp.current?.state.value),
          "chlorine": this.state.cl,
          "quality": this.state.waterQuality,
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

  private valid(): boolean {
    this.setState({clValid: false});

    if (this.state.cl >= 0 && this.state.cl <= 5) {
      this.setState({clValid: true});

      return true;
    }

    return false;
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
          <DialogTitle>Obtener muestra</DialogTitle>
          <DialogContent>
            <DialogContentText id="alert-dialog-slide-description">
              En base a las métricas obtenidas por el micro-controlador,
              rellene los valores del cloro medidos por usted
              y categorice la calidad del agua.
            </DialogContentText>
            <Stack marginTop="20px" spacing={2}>
              <TextField
                id="chlorine"
                sx={{marginTop:"10px"}}
                label="Cloro (mg/L)"
                type="number"
                value={this.state.cl}
                onChange={event =>
                    this.setState({cl: Number(event.target.value)})}
                InputProps={{ inputProps: { min: 0, max: 5 } }}                  
                size="medium"
                error={!this.state.clValid}
                helperText="Introduzca un valor entre 0 y 5 mg/L"/> 
              <InputLabel variant="standard" htmlFor="waterQuality">
                Calidad del agua
              </InputLabel>                                 
              <Select
                  id="waterQuality"
                  sx={{marginTop:"10px"}}
                  label="Calidad del agua"
                  value={this.state.waterQuality}
                  onChange={event =>
                      this.setState(
                      {waterQuality: Number(event.target.value)})}
                  size="small"
              >        
                <MenuItem value={0}>Mala</MenuItem>
                <MenuItem value={1}>Regular</MenuItem>
                <MenuItem value={2}>Buena</MenuItem>        
              </Select>
          </Stack>
            <Stack
              marginTop={"20px"}
              direction={{ xs: 'column', sm: 'row' }}
              spacing={2}
            >
              <Grid container spacing={2}>
                <Grid item xs={12} md={2} lg={2}>
                  <Item>
                    <Meassure 
                      ref={this.meassurePh} name={literals.phName} 
                      unitName={literals.phUnit} src=""/>
                  </Item>
                </Grid>
                <Grid item xs={12} md={5} lg={5}>
                  <Item>
                    <Meassure 
                      ref={this.meassureOrp} name={literals.orpName} 
                      unitName={literals.orpUnit} src=""/>
                  </Item>
                </Grid>
                <Grid item xs={12} md={5} lg={5}>
                  <Item>
                    <Meassure 
                      ref={this.meassureTemp} name={literals.temperatureName} 
                      unitName={literals.temperatureUnit} src=""/>
                  </Item>
                </Grid>
              </Grid>
            </Stack>
          </DialogContent>
          <DialogActions>
              <Box sx={{ m: 1, position: 'relative' }}>
                  <Button onClick={
                    () => this.close(true)}>
                      {this.state.saving ? "Enviando" : "Enviar"}</Button>
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
              <Button
               disabled={this.state.saving}
               onClick={() => this.close(false)}>Cancelar
              </Button>
          </DialogActions>
      </Dialog>
    </div>
    );
  }
}

