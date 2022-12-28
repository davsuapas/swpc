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

import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Chart from './chart';
import Meassure from './meassure';
import AppBar from '@mui/material/AppBar';
import ExitToAppIcon from '@mui/icons-material/ExitToApp'
import { Assignment } from '@mui/icons-material';
import Config from '../config/config';
import Tooltip from '@mui/material/Tooltip';
import { CircularProgress } from '@mui/material';
import React from 'react';
import Alert from '../support/alert';
import { colorPurple } from '../support/color';
import SocketFactory, { Metrics } from '../net/socket';
import { Websocket } from 'websocket-ts/lib';
import { Navigate } from 'react-router-dom';
import { MediaQuery, MediaQueryAPI } from '../support/mediaquery';
import User from '../login/user';

const drawerWidth: number = 255;

const temperatureName = "TEMPERATURA";
const temperatureUnit = "grados";

const phName = "pH";
const phUnit = "";

const chlorineName = "CLORO";
const chlorineUnit = "mg/L";

const mdTheme = createTheme();

interface DashboardState{
  loadingConfig: boolean;
  standby: boolean;
  shutdown: boolean;
}

export interface Actions {
  activeStandby(active: boolean): void;
  activeLoadingConfig(active: boolean): void;
  shutdown(): void;
}

export default class Dashboard extends React.Component<any, DashboardState> implements Actions {

  private socket: Websocket;

  private media: React.RefObject<MediaQueryAPI>

  private config: React.RefObject<Config>;
  private alert: React.RefObject<Alert>;

  private chartTemp: React.RefObject<Chart>;
  private chartPh: React.RefObject<Chart>;
  private chartCl: React.RefObject<Chart>;

  private meassureTemp: React.RefObject<Meassure>;
  private meassurePh: React.RefObject<Meassure>;
  private meassureCl: React.RefObject<Meassure>;

  private user: User

  constructor(props: any) {
    super(props);

    this.state = {
      loadingConfig: false,
      standby: true,
      shutdown: false
    };

    this.media = React.createRef();

    this.config = React.createRef<Config>();
    this.alert = React.createRef<Alert>();

    this.chartTemp = React.createRef<Chart>();
    this.chartPh = React.createRef<Chart>();
    this.chartCl = React.createRef<Chart>();

    this.meassureTemp = React.createRef<Meassure>();
    this.meassurePh = React.createRef<Meassure>();
    this.meassureCl = React.createRef<Meassure>();

    this.user = new User(this);

    const sfactory = new SocketFactory(this.alert, this);
    sfactory.event.streamMetrics = this.streamMetrics.bind(this);

    this.socket = sfactory.open();
  }

  activeStandby(active: boolean): void {
    this.setState({standby: active});
  }

  activeLoadingConfig(active: boolean): void {
    this.setState({loadingConfig: active});
  }

  shutdown(): void {
    this.setState({shutdown: true});
  }

  // streamMetrics sends all the metrics received by socket to all the chart controls
  private streamMetrics(metrics: Metrics) {
    this.chartTemp.current?.setData(metrics.temp);
    this.chartPh.current?.setData(metrics.ph);
    this.chartCl.current?.setData(metrics.chlorine);
  }

  private exit() {
    this.socket.close();
    this.user.logoff();
  }

  componentDidMount(): void {
    // Pipeline between the chart and meassure. it sends the last meassure received by chart to the meassure control
    if (this.chartTemp.current) {
      this.chartTemp.current.event.lastDataReceived = (m) => {this.meassureTemp.current?.setMeassure(m);}
    }
    if (this.chartPh.current) {
      this.chartPh.current.event.lastDataReceived = (m) => {this.meassurePh.current?.setMeassure(m);}
    }
    if (this.chartCl.current) {
      this.chartCl.current.event.lastDataReceived = (m) => {this.meassureCl.current?.setMeassure(m.toFixed(2));}
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
                <Tooltip title="Configuración">
                  <IconButton color="inherit" onClick={() => this.config.current?.open(this)}>
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
                <Tooltip title="Salir">
                  <IconButton color="inherit" onClick={() => this.exit()}>
                      <ExitToAppIcon />
                  </IconButton>
                </Tooltip>
              </Toolbar>
            </AppBar>
              <Container maxWidth="xl" sx={{ mt:5, mb: 5, position: "relative"}}>
                <Grid container spacing={2}>
                  <Grid item xs={12} md={9} lg={10}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Chart ref={this.chartTemp} name={temperatureName} 
                        unitName={temperatureUnit} theme={mdTheme} media={this.media.current} />
                    </Paper>
                  </Grid>
                  <Grid item xs={12} md={3} lg={2}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Meassure ref={this.meassureTemp} name={temperatureName} unitName={temperatureUnit} src="temp.png" />
                    </Paper>
                  </Grid>
                  <Grid item xs={12} md={9} lg={10}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Chart ref={this.chartPh} name={phName}
                        unitName={phUnit} theme={mdTheme} media={this.media.current} />
                    </Paper>
                  </Grid>
                  <Grid item xs={12} md={3} lg={2}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Meassure ref={this.meassurePh} name={phName} unitName={phUnit} src="ph.png" />
                    </Paper>
                  </Grid>
                  <Grid item xs={12} md={9} lg={10}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Chart ref={this.chartCl} name={chlorineName}
                        unitName={chlorineUnit} theme={mdTheme} media={this.media.current} />
                    </Paper>
                  </Grid>
                  <Grid item xs={12} md={3} lg={2}>
                    <Paper
                      sx={{
                        p: 2,
                        display: 'flex',
                        flexDirection: 'column',
                        height: drawerWidth,
                      }}
                    >
                      <Meassure ref={this.meassureCl} name={chlorineName} unitName={chlorineUnit} src="chlorine.png" />
                    </Paper>
                  </Grid>
                </Grid>
                <div style={{position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)"}}>
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
            <Config ref={this.config} alert={this.alert} actions={this}/>
            {this.state.shutdown && (<Navigate to="/" replace={true}/>)}
            <MediaQuery ref={this.media} theme={mdTheme}></MediaQuery>
      </ThemeProvider>
    );    
  }
}
