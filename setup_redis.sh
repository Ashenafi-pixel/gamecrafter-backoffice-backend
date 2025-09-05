# Redis Installation and Configuration
sudo apt update
sudo apt install -y redis-server

# Configure Redis
sudo sed -i 's/^# requirepass foobared/requirepass tucanbit_redis_password/' /etc/redis/redis.conf
sudo sed -i 's/^bind 127.0.0.1 ::1/bind 0.0.0.0/' /etc/redis/redis.conf
sudo sed -i 's/^# maxmemory <bytes>/maxmemory 256mb/' /etc/redis/redis.conf
sudo sed -i 's/^# maxmemory-policy noeviction/maxmemory-policy allkeys-lru/' /etc/redis/redis.conf

# Start and enable Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server
sudo systemctl status redis-server
