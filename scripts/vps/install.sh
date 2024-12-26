#!/bin/bash

# Install swpc for Virtual Private Server

# Create swpc user
if ! id "swpc" &>/dev/null; then
  useradd -r -s /bin/false swpc
fi

# Remove existing directories if they exist
rm -rf /opt/swpc /var/log/swpc /var/lib/swpc

# Create directories
mkdir -p /opt/swpc /var/log/swpc /var/lib/swpc /etc/swpc

# Copy binary files to /opt/swpc
cp -r ./bin /opt/swpc/
cp ./swpc.env /etc/swpc/swpc.env

# Set ownership and permissions
chown -R swpc:swpc /opt/swpc /var/log/swpc /var/lib/swpc /etc/swpc
chmod -R 750 /opt/swpc
chmod -R 400 /etc/swpc
chmod -R 600 /var/log/swpc
chmod -R 700 /var/lib/swpc

./update-service.sh

echo "SWPC installed successfully"