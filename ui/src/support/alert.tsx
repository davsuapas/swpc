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

export default forwardRef( (props: any, ref: any) => {
    const [openv, setOpenv] = React.useState(false);

    const [title, setTitle] = React.useState("");
    const [contentv, setContentv] = React.useState("");

    const open = () => {
      setOpenv(true);
    };
  
    const close = () => {
      setOpenv(false);
    };

    const content = (title: string, content: string) => {
      setTitle(title);
      setContentv(content);
    }

    // Export the function
    if (ref) ref.current = {open, content}

    return (
    <div>
        <Dialog
            open={openv}
            TransitionComponent={Transition}
            keepMounted
            onClose={close}
        >
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <DialogContentText id="alert-dialog-slide-description">
                    {contentv}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={close}>Entendido</Button>
            </DialogActions>
        </Dialog>
    </div>
    );
});