#!/bin/bash

# Build the SPA ui for debugging

PATH_PUBLIC=./cmd/swpc-server/public

cd ui
npm run build
cd ..
rm -R $PATH_PUBLIC
cp -R ./ui/build $PATH_PUBLIC