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

function preventDefault(event: React.MouseEvent) {
  event.preventDefault();
}

interface MeassureProps {
  name: string;
  unitName: string;
  src: string;
}

interface MeassureState {
  value: number;
}

export default class Meassure extends React.Component<MeassureProps, MeassureState> {

  constructor(props: MeassureProps) {
    super(props)

    this.state = {
      value: 0
    }
  }

  setMeassure(value: number) {
    this.setState({
      value: value
    });
  }

  render(): React.ReactNode {
    return (
      <React.Fragment>
        <Title>{this.props.name}</Title>
        <Typography component="p" variant="h4" sx={{marginTop: '25px', marginBottom: "20px"}}>
          {this.state.value} {this.props.unitName}
        </Typography>
        {this.props.src != "" && (
          <img src={this.props.src} width="80" height="80"></img>
        )}
      </React.Fragment>
    );
  }
}