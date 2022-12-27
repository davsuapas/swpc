import { useMediaQuery } from "@mui/material";
import { forwardRef, useEffect, useImperativeHandle, useState } from "react";

export interface MediaQueryAPI {
    isMd: () => boolean
    isLg: () => boolean
  }
  
// MediaQuery gets information about media
export const MediaQuery = forwardRef<MediaQueryAPI, any>((props, ref) => {
    const isMd = useMediaQuery(props.theme.breakpoints.up('md'));
    const isLg = useMediaQuery(props.theme.breakpoints.up('lg'));
    
    useImperativeHandle(ref, () => ({
        isMd: () => {return !isLg && isMd},
        isLg: () => {return isLg && isMd}
    }));

    return (<div/>);
});
  