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

import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Slide from '@mui/material/Slide';
import { TransitionProps } from '@mui/material/transitions';
import React, { forwardRef } from 'react';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="up" ref={ref} {...props} />;
});

interface AlertState {
  title: string,
  content: string,
  open: boolean
}

interface AlertEvents {
  closed: () => void
}

// Alert creates a component to display system alerts
export default class Alert extends React.Component<any, AlertState> {

  events: AlertEvents

  constructor(props: any) {
    super(props);
    this.state = {
      title: "",
      content: "",
      open: false
    };
    this.events = {
      closed: () => {}
    };
  }

  // content updates title and content properties
  content(title: string, content: string) {
    this.setState({
      title: title,
      content: content
    });
  }

  // open opens the alert form
  open() {
    this.setState({open: true});
  }

  private close() {
    this.setState({open: false});
    this.events.closed();
    this.iniClosedEvent();
  }

  private iniClosedEvent() {
    this.events.closed = () => {};
  }

  render(): React.ReactNode {
    return (
      <div>
          <Dialog
              open={this.state.open}
              TransitionComponent={Transition}
              keepMounted
              onClose={this.close.bind(this)}
          >
              <DialogTitle>{this.state.title}</DialogTitle>
              <DialogContent>
                  <DialogContentText id="alert-dialog-slide-description">
                      {this.state.content}
                  </DialogContentText>
              </DialogContent>
              <DialogActions>
                  <Button onClick={this.close.bind(this)}>Entendido</Button>
              </DialogActions>
          </Dialog>
      </div>
      );    
  }
}