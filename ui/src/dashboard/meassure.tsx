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

import * as React from 'react';
import Typography from '@mui/material/Typography';
import Title from './title';

export const defaulBoxColor = "#ffffff"

const styleWarn = {
  color: "#ff8c00"
};

interface MeassureProps {
  name: string;
  unitName: string;
  src: string;
}

interface MeassureState {
  value: string;
  warn: boolean;
  boxColor: string;
}

export default class Meassure extends React.Component<MeassureProps, MeassureState> {

  constructor(props: MeassureProps) {
    super(props)

    this.state = {
      value: "",
      warn: false,
      boxColor: defaulBoxColor
    }
  }

  setMeassure(value: string, boxColor = defaulBoxColor) {
    this.setState({
      value: value,
      warn: false,
      boxColor: boxColor
    });
  }

  setWarn(value: string) {
    this.setState({
      value: value,
      warn: true,
      boxColor: defaulBoxColor
    });
  }

  render(): React.ReactNode {
    return (
      <div style={{ backgroundColor: this.state.boxColor}}>
        <React.Fragment>
          <Title>{this.props.name}</Title>
          {this.state.warn && (
            <Typography component="p" variant="h6" style={styleWarn} sx={{marginTop: '25px', marginBottom: "20px"}}>
              {this.state.value}
            </Typography>
          )}
          {!this.state.warn && (
            <Typography component="p" variant="h4" sx={{marginTop: '25px', marginBottom: "20px"}}>
              {this.state.value} {this.props.unitName}
            </Typography>
          )}
          {this.props.src != "" && (
            <img src={this.props.src} width="80" height="80"></img>
          )}
        </React.Fragment>
      </div>
    );
  }
}