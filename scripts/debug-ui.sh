#!/bin/bash

# Build the spa ui for debugging

PATH_PUBLIC=./cmd/swpoolcontroller-server/public

cd ui
npm run build
cd ..
rm -R $PATH_PUBLIC
cp -R ./ui/build $PATH_PUBLIC