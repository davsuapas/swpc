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

import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Chart from './chart';
import Meassure, { defaulBoxColor } from './meassure';
import AppBar from '@mui/material/AppBar';
import ExitToAppIcon from '@mui/icons-material/ExitToApp'
import RefreshIcon from '@mui/icons-material/Refresh';
import { AppRegistration, Assignment } from '@mui/icons-material';
import Config from '../config/config';
import Tooltip from '@mui/material/Tooltip';
import { CircularProgress } from '@mui/material';
import React from 'react';
import Alert from '../info/alert';
import { colorPurple } from '../support/color';
import SocketFactory, { CommStatus, Metrics } from '../net/socket';
import { Websocket } from 'websocket-ts/lib';
import { MediaQuery, MediaQueryAPI } from '../support/mediaquery';
import { logoff } from '../auth/user';
import * as literals from '../support/literals';
import Sample from '../ai/sample';
import { appConfig } from '../app/config';
import Fetch from '../net/fetch';

const drawerWidth: number = 255;

const mdTheme = createTheme();

interface DashboardState {
  loadingConfig: boolean;
  standby: boolean;
  refresh: boolean;
  wqBoxColor: string;
}

export interface Actions {
  activeStandby(active: boolean): void;
  activeLoadingConfig(active: boolean): void;
}

export default class Dashboard extends React.Component<any, DashboardState> implements Actions {

  private configUI: boolean;
  private sampleUI: boolean;

  private sfactory: SocketFactory;
  private socket: Websocket;

  private fetch: Fetch;

  private media: React.RefObject<MediaQueryAPI>

  private config: React.RefObject<Config>;
  private sample: React.RefObject<Sample>;
  private alert: React.RefObject<Alert>;

  private chartTemp: React.RefObject<Chart>;
  private chartPh: React.RefObject<Chart>;
  private chartOrp: React.RefObject<Chart>;

  private meassureWq: React.RefObject<Meassure>;
  private meassureCl: React.RefObject<Meassure>;

  private meassureTemp: React.RefObject<Meassure>;
  private meassurePh: React.RefObject<Meassure>;
  private meassureOrp: React.RefObject<Meassure>;

  // Controls that the first time metrics are sent it
  // requests the data on demand. 
  // After that it will be the user who requests them
  private ondemandData: boolean;

  constructor(props: any) {
    super(props);

    this.state = {
      loadingConfig: false,
      standby: true,
      refresh: false,
      wqBoxColor: defaulBoxColor
    };

    const config = appConfig()
    this.configUI = config.iotConfig;
    this.sampleUI = config.aiSample

    this.media = React.createRef();

    this.config = React.createRef<Config>();
    this.sample = React.createRef<Sample>();
    this.alert = React.createRef<Alert>();

    this.chartTemp = React.createRef<Chart>();
    this.chartPh = React.createRef<Chart>();
    this.chartOrp = React.createRef<Chart>();

    this.meassureWq = React.createRef<Meassure>();
    this.meassureCl = React.createRef<Meassure>();

    this.meassureTemp = React.createRef<Meassure>();
    this.meassurePh = React.createRef<Meassure>();
    this.meassureOrp = React.createRef<Meassure>();

    this.sfactory = new SocketFactory(this.alert, this);
    this.sfactory.event.streamMetrics = this.streamMetrics.bind(this);
    this.sfactory.event.status = this.socketStatus.bind(this);

    this.socket = this.sfactory.open();

    this.fetch = new Fetch(this.alert);

    this.ondemandData = false;
  }

  activeStandby(active: boolean): void {
    this.setState({ standby: active });
  }

  activeLoadingConfig(active: boolean): void {
    this.setState({ loadingConfig: active });
  }

  private loadOndemandData(metrics: Metrics | null): void {
    if (this.sfactory.state != CommStatus.broadcasting) {
      this.alert.current?.content(
        "No se detectan métricas",
        "Sin métricas es imposible predecir ni la calidad del agua " +
        "ni el cloro. Espere a que el micro-controlador envíe las métricas");
      this.alert.current?.open();

      return
    }

    this.setState({ refresh: true });

    let meassure = {}

    if (metrics == null) {
      meassure = {
        temp: this.meassureTemp.current?.state.value.toString(),
        ph: this.meassurePh.current?.state.value.toString(),
        orp: this.meassureOrp.current?.state.value.toString(),
      };
    } else {
      meassure = {
        temp: metrics.temp[0].toString(),
        ph: metrics.ph[0].toString(),
        orp: metrics.orp[0].toString(),
      };
    }

    this.fetch?.send("/api/web/predict", {
      method: "POST",
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(meassure)
    },
      async (result: Response) => {
        let manejado = true;

        if (result.ok) {
          const res = await result.json();

          let boxColor = defaulBoxColor;
          let wq = "";
          switch (res.wq) {
            case "bad":
              wq = "MALA";
              boxColor = "#EC7063";
              break;
            case "regular":
              wq = "REGULAR";
              boxColor = "#F4D03F";
              break;
            default:
              wq = "BUENA";
              boxColor = "#5DADE2";
              break;
          }

          this.setState({ wqBoxColor: boxColor });
          this.meassureWq.current?.setMeassure(wq, boxColor);

          this.meassureCl.current?.setMeassure(Number(res.cl).toFixed(2).toString());
        } else {
          if (result.status == 404) {
            const message = "No hay suficientes datos para predecir";
            this.meassureWq.current?.setWarn(message);
            this.meassureCl.current?.setWarn(message);

            this.alert.current?.content(
              "Modelo predictivo inexistente",
              "Todavía no existen sufificientes muestras para predecir " +
              "ni la calidad del agua ni el cloro. Continue realizando " +
              "muestras");
            this.alert.current?.open();
          } else {
            manejado = false;
          }
        }

        this.setState({ refresh: false });

        return manejado;
      },
      () => {
        this.setState({ refresh: false });
      });
  }

  private socketStatus(status: CommStatus) {
    if (status == CommStatus.broadcasting) {
      this.ondemandData = true;
    }
  }

  private requestOndemandData(metrics: Metrics) {
    if (this.ondemandData) {
      this.loadOndemandData(metrics);
    }

    this.ondemandData = false;
  }

  // streamMetrics sends all the metrics received by socket to all the chart controls
  private streamMetrics(metrics: Metrics) {
    this.chartTemp.current?.setData(metrics.temp);
    this.chartPh.current?.setData(metrics.ph);
    this.chartOrp.current?.setData(metrics.orp);

    this.requestOndemandData(metrics);
  }

  private exit() {
    this.socket.close();
    logoff();
  }

  componentDidMount(): void {
    // Pipeline between the chart and meassure. 
    // it sends the last meassure received by chart to the meassure control
    if (this.chartTemp.current) {
      this.chartTemp.current.event.lastDataReceived = (m) => { this.meassureTemp.current?.setMeassure(m); }
    }
    if (this.chartPh.current) {
      this.chartPh.current.event.lastDataReceived = (m) => { this.meassurePh.current?.setMeassure(m); }
    }
    if (this.chartOrp.current) {
      this.chartOrp.current.event.lastDataReceived = (m) => { this.meassureOrp.current?.setMeassure(m.toFixed(2)); }
    }
  }

  render(): React.ReactNode {
    return (
      <ThemeProvider theme={mdTheme}>
        <CssBaseline />
        <AppBar position="static">
          <Toolbar>
            <Typography
              component="h1"
              variant="h6"
              color="inherit"
              noWrap
              sx={{ flexGrow: 1 }}
            >
              Métricas piscina
            </Typography>
            <Tooltip title="Refescar las prediciones de la calidad del agua y el cloro">
              <IconButton
                color="inherit"
                onClick={() => this.loadOndemandData(null)}>
                <RefreshIcon />
                {this.state.refresh && (
                  <CircularProgress
                    size={40}
                    sx={{
                      color: colorPurple,
                      position: 'absolute',
                      zIndex: 1,
                    }}
                  />
                )}
              </IconButton>
            </Tooltip>
            {this.sampleUI && (
              <Tooltip title="Muestra">
                <IconButton
                  color="inherit"
                  onClick={
                    () => {
                      if (this.sfactory.state == CommStatus.broadcasting &&
                        this.meassureTemp.current != undefined &&
                        this.meassurePh.current != undefined &&
                        this.meassureOrp.current != undefined
                      ) {
                        this.sample.current?.open(
                          this.meassureTemp.current.state.value,
                          this.meassurePh.current.state.value,
                          this.meassureOrp.current.state.value);
                      } else {
                        this.alert.current?.content(
                          "No se detectan métricas",
                          "No se puede realizar una muestra sino se " +
                          "envían métricas desde el micro-controlador");
                        this.alert.current?.open();
                      }
                    }
                  }>
                  <AppRegistration />
                </IconButton>
              </Tooltip>
            )}
            {this.configUI && (
              <Tooltip title="Configuración">
                <IconButton
                  color="inherit"
                  onClick={() => this.config.current?.open(this)}>
                  <Assignment />
                  {this.state.loadingConfig && (
                    <CircularProgress
                      size={40}
                      sx={{
                        color: colorPurple,
                        position: 'absolute',
                        zIndex: 1,
                      }}
                    />
                  )}
                </IconButton>
              </Tooltip>
            )}
            <Tooltip title="Salir">
              <IconButton color="inherit" onClick={() => this.exit()}>
                <ExitToAppIcon />
              </IconButton>
            </Tooltip>
          </Toolbar>
        </AppBar>
        <Container maxWidth="xl" sx={{ mt: 5, mb: 5, position: "relative" }}>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6} lg={2}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                  backgroundColor: this.state.wqBoxColor
                }}
              >
                <Meassure ref={this.meassureWq} name={literals.wqName} unitName={literals.wqUnit} src="w.png" />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={2}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Meassure ref={this.meassureCl} name={literals.clName} unitName={literals.clUnit} src="cl.png" />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={2}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Meassure ref={this.meassurePh} name={literals.phName} unitName={literals.phUnit} src="ph.png" />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={3}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Meassure ref={this.meassureOrp} name={literals.orpName} unitName={literals.orpUnit} src="orp.png" />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={3}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Meassure ref={this.meassureTemp} name={literals.temperatureName} unitName={literals.temperatureUnit} src="temp.png" />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={12}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Chart ref={this.chartPh} name={literals.phName}
                  unitName={literals.phUnit} theme={mdTheme} media={this.media.current} />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={6}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Chart ref={this.chartOrp} name={literals.orpName}
                  unitName={literals.orpUnit} theme={mdTheme} media={this.media.current} />
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} lg={6}>
              <Paper
                sx={{
                  p: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  height: drawerWidth,
                }}
              >
                <Chart ref={this.chartTemp} name={literals.temperatureName}
                  unitName={literals.temperatureUnit} theme={mdTheme} media={this.media.current} />
              </Paper>
            </Grid>
          </Grid>
          <div style={{ position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)" }}>
            {this.state.standby && (
              <CircularProgress
                size={80}
                sx={{
                  color: colorPurple,
                  position: 'absolute',
                  zIndex: 1,
                  margin: "0 auto"
                }}
              />
            )}
          </div>
        </Container>
        <Alert ref={this.alert}></Alert>
        <Config ref={this.config} alert={this.alert} />
        <Sample ref={this.sample} alert={this.alert} />
        <MediaQuery ref={this.media} theme={mdTheme}></MediaQuery>
      </ThemeProvider>
    );
  }
}
