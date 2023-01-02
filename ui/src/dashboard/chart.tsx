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
import TaskInterval from '../support/interval';
import { Theme } from '@mui/material/styles';
import { MediaQueryAPI } from '../support/mediaquery';

interface ChartEvent {
  lastDataReceived: (lastData: any) => void;
}

interface ChartProps {
  name: string;
  unitName: string;
  theme: Theme;
  media: MediaQueryAPI | null
}

interface ChartItem {
  time: string,
  amount: number
}

interface ChartState {
  data: ChartItem[];
}

export default class Chart extends React.Component<ChartProps, ChartState> {

  event: ChartEvent;

  private model: Model;

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
  private index: number;
  private cumuSize: number;

  constructor(private chart: Chart) {
    this.buffer = [];
    this.index = -1; // Empty stack
    this.cumuSize = 0;

    this.task = new TaskInterval(intervalInSeconds * 1000, this.readBuffer.bind(this));
  }

  // setData updates the buffer and start reading
  setData(data: any[]) {
    if (data.length == 0) {
      return
    }

    this.cumuSize += data.length;

    // In order to always display the most current data,
    // data that has not yet been read is added to the top of the stack and inserted as a buffer,
    // with a maximum of twice the amount of data being sent.
    if (this.hasBuffer()) {
      data.push(...this.buffer.slice(this.index, this.index + data.length));
    }
    this.buffer = data;
    this.index = 0;

    this.task.start();
  }

  // stop stops the reader
  stop() {
    this.task.stop();
  }

  // readBuffer reads the buffer until the end and updating then chart
  private readBuffer(): boolean {
    if (this.hasBuffer()) {
      if (this.cumuSize % 10 == 0) {
        console.log("Chart (warning) -> More metrics are received than extracted: " + this.cumuSize +
         ". Tamaño buffer: " + this.buffer.length);
      }

      const date = new Date();
      const formattedTime = `${date.getHours()}:${date.getMinutes()}:${date.getSeconds()}`;
      const amount = this.pop(); 
 
      // raises last data received event
      this.chart.lastDataReceived(amount);

      this.chart.setState(prevState => {
        return this.setChartState(prevState, {
            time: formattedTime,
            amount: amount
          });
      });

      if (this.hasBuffer()) {
        return false; // Continue extracting with the next each interval
      }
    }

    return true; // Cancel, because there is nothing more to extract
  }

  // setChartState sets the chart state adding the new metric,
  // depending on the size of the screen, more or less data is displayed
  private setChartState(prevState: ChartState, item: ChartItem): ChartState {
    const sizeXs = 3;
    const sizeMd = 7;
    const sizeLg = 11;

    let size = sizeXs;
    if (this.chart.props.media?.isMd()) {
      size = sizeMd;
    } else if (this.chart.props.media?.isLg()) {
      size = sizeLg;
    }

    if (prevState.data.length > size) {
      prevState.data.splice(0, prevState.data.length - size);
    }

    return {data: [...prevState.data, item]};
  }

  private hasBuffer(): boolean {
    return this.index > -1;
  }

  // pop gets the first element of the buffer
  private pop() {
    const data = this.buffer[this.index++];
    if (this.buffer.length == this.index) {
      this.index = -1;
    }
    this.cumuSize--;

    return data;
  }
}