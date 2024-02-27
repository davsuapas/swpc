#!/bin/bash

# Build the SPA ui for debugging

path_web=$1
omap=$2

if [[ -z "$path_web" ]]; then
    path_web="./cmd/swpc-server"
fi

path_web="$path_web/public"

cd ui
npm run build
cd ..
rm -R "$path_web"
if [[ "$omap" == "-no-map" ]]; then
  rm ./ui/build/static/js/*.map
fi
cp -R ./ui/build "$path_web"
