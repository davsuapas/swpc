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
import { LineChart, Line, XAxis, YAxis, Label, ResponsiveContainer, CartesianGrid } from 'recharts';
import Title from './title';
import { Theme } from '@mui/material/styles';
import TaskInterval from '../support/interval';

interface ChartEvent {
  lastDataReceived: (lastData: any) => void;
}

interface ChartProps {
  name: string;
  unitName: string;
  theme: Theme;
}

interface ChartState {
  data: any;
}

export default class Chart extends React.Component<ChartProps, ChartState> {

  private model: Model;

  private event: ChartEvent;

  constructor(props: ChartProps) {
    super(props);

    this.state = {
      data: []
    };

    this.event = {
      lastDataReceived: (data) => {}
    };

    this.model = new Model(this);
  }

  // setData updates the buffer of the chart
  setData(data: any[]) {
    this.model.setData(data);
  }

  componentWillUnmount(): void {
    this.model.stop()
  }

  // lastDataReceived raises last data received event
  lastDataReceived(data: any) {
    this.event.lastDataReceived(data);
  }

  render(): React.ReactNode {
    return (
      <React.Fragment>
        <Title>{this.props.name}</Title>
        <ResponsiveContainer>
          <LineChart
            data={this.state.data}
            margin={{
              top: 16,
              right: 16,
              bottom: 0,
              left: 24,
            }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="time"
              stroke={this.props.theme.palette.text.secondary}
              style={this.props.theme.typography.body2}
            />
            <YAxis
              stroke={this.props.theme.palette.text.secondary}
              style={this.props.theme.typography.body2}
            >
              <Label
                angle={270}
                position="left"
                style={{
                  textAnchor: 'middle',
                  fill: this.props.theme.palette.text.primary,
                  ...this.props.theme.typography.body1,
                }}
              >
                {this.props.unitName}
              </Label>
            </YAxis>
            <Line
              isAnimationActive={false}
              type="monotone"
              dataKey="amount"
              stroke={this.props.theme.palette.primary.main}
              dot={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </React.Fragment>
    );
  }
}

const intervalInSeconds = 1;

// Model nanages the chart data model
class Model {

  private task: TaskInterval;

  private buffer: any[];

  constructor(private chart: Chart) {
    this.buffer = [];
    this.task = new TaskInterval(intervalInSeconds * 1000, this.readBuffer);
  }

  // setData updates the buffer and start reading
  setData(data: any[]) {
    const read = !this.hasBuffer() && data.length > 0;

    this.buffer.push(data);

    if (read) {
      this.task.start();
    }
  }

  // stop stops the reader
  stop() {
    this.task.stop();
  }

  // readBuffer reads the buffer until the end and updating then chart
  private readBuffer(): boolean {
    if (this.hasBuffer()) {
      const date = new Date();
      const formattedTime = `${date.getHours()}:${date.getMinutes()}:${date.getSeconds()}`;

      const amount = this.pop(); 

      // raises last data received event
      this.chart.lastDataReceived(amount);

      this.chart.setState(prevState => {
        return {
          data: [...prevState.data, {
            time: formattedTime,
            amount: amount
          }]
        };
      });
      if (this.hasBuffer()) {
        return false; // Continue extracting with the next each interval
      }
    }

    return true; // Cancel
  }

  private hasBuffer(): boolean {
    return this.buffer.length > 0;
  }

  // pop gets the first element of the buffer and it's remove it.
  private pop() {
    return this.buffer.shift();
  }
}