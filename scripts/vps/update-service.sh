#!/bin/bash

# Update swpc service for Virtual Private Server

# Update systemd service file
cat <<EOF >/etc/systemd/system/swpc.service
[Unit]
Description=SWPC Service
After=network.target

[Service]
Type=simple
User=swpc
EnvironmentFile=/etc/swpc/swpc.env
WorkingDirectory=/opt/swpc/bin
ExecStart=/opt/swpc/bin/swpc-server
StandardOutput=append:/var/log/swpc/swpc.log
StandardError=append:/var/log/swpc/swpc.log
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd, enable and start the service if it does not exist, otherwise reload it
if systemctl list-units --full -all | grep -Fq 'swpc.service'; then
  systemctl daemon-reload
  systemctl restart swpc.service
else
  systemctl daemon-reload
  systemctl enable swpc.service
  systemctl start swpc.service
fi

systemctl status swpc.service

echo "SWPC service updated successfully"
