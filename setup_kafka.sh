# Apache Kafka Installation and Configuration
# Install Java (required for Kafka)
sudo apt install -y openjdk-17-jdk

# Create kafka user
sudo useradd -r -s /bin/false kafka

# Download and install Kafka
cd /opt
sudo wget https://downloads.apache.org/kafka/2.13-3.7.0/kafka_2.13-3.7.0.tgz
sudo tar -xzf kafka_2.13-3.7.0.tgz
sudo mv kafka_2.13-3.7.0 kafka
sudo chown -R kafka:kafka /opt/kafka

# Create systemd service for Zookeeper
sudo tee /etc/systemd/system/zookeeper.service > /dev/null << 'EOF'
[Unit]
Description=Apache Zookeeper server
Documentation=http://zookeeper.apache.org
Requires=network.target remote-fs.target
After=network.target remote-fs.target

[Service]
Type=simple
User=kafka
Group=kafka
ExecStart=/opt/kafka/bin/zookeeper-server-start.sh /opt/kafka/config/zookeeper.properties
ExecStop=/opt/kafka/bin/zookeeper-server-stop.sh
Restart=on-abnormal

[Install]
WantedBy=multi-user.target
EOF

# Create systemd service for Kafka
sudo tee /etc/systemd/system/kafka.service > /dev/null << 'EOF'
[Unit]
Description=Apache Kafka server
Documentation=http://kafka.apache.org/documentation.html
Requires=zookeeper.service
After=zookeeper.service

[Service]
Type=simple
User=kafka
Group=kafka
ExecStart=/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties
ExecStop=/opt/kafka/bin/kafka-server-stop.sh
Restart=on-abnormal

[Install]
WantedBy=multi-user.target
EOF

# Configure Kafka
sudo sed -i 's/#listeners=PLAINTEXT:\/\/:9092/listeners=PLAINTEXT:\/\/0.0.0.0:9092/' /opt/kafka/config/server.properties
sudo sed -i 's/#advertised.listeners=PLAINTEXT:\/\/your.host.name:9092/advertised.listeners=PLAINTEXT:\/\/51.21.181.162:9092/' /opt/kafka/config/server.properties

# Start services
sudo systemctl daemon-reload
sudo systemctl start zookeeper
sudo systemctl start kafka
sudo systemctl enable zookeeper
sudo systemctl enable kafka

# Check status
sudo systemctl status zookeeper
sudo systemctl status kafka
