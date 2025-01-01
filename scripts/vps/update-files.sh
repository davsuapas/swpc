#!/bin/bash

# Update files for Virtual Private Server

# Remove existing directories if they exist
rm -rf /opt/swpc/*

# Copy binary files to /opt/swpc
cp -r ./bin /opt/swpc/
cp ./swpc.env /etc/swpc/swpc.env

# Remove log file if they exist
rm /var/log/swpc/swpc.log

./update-service.sh

echo "SWPC files updated successfully"

echo "Eliminar vps/ y swpc.zip"