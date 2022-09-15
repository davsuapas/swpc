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

import * as React from 'react';
import Typography from '@mui/material/Typography';
import Title from './title';

function preventDefault(event: React.MouseEvent) {
  event.preventDefault();
}

interface MeassureProps {
  name: string;
  value: string;
  unitName: string;
}

export default function Meassure(props: MeassureProps) {
  return (
    <React.Fragment>
      <Title>{props.name}</Title>
      <Typography component="p" variant="h4">
        <br/>
        {props.value} {props.unitName}
      </Typography>
      <Typography sx={{ mb: 1.5 }} color="text.secondary">
        <br/>
        El indicador mide el valor actual que proporciona la piscina a través del controlador
      </Typography>
    </React.Fragment>
  );
}
