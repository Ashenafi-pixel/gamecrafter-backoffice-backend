# TucanBIT Systemd Service
sudo tee /etc/systemd/system/tucanbit.service > /dev/null << 'EOF'
[Unit]
Description=TucanBIT Online Casino API
After=network.target postgresql.service redis-server.service kafka.service
Wants=postgresql.service redis-server.service kafka.service

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/opt/tucanbit
ExecStart=/opt/tucanbit/tucanbit
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/opt/tucanbit/config/production.yaml

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable tucanbit
sudo systemctl start tucanbit
sudo systemctl status tucanbit
