# PostgreSQL Installation and Configuration
# Install PostgreSQL
sudo apt install -y postgresql postgresql-contrib

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql << 'EOF'
CREATE DATABASE tucanbit;
CREATE USER tucanbit WITH PASSWORD '5kj0YmV5FKKpU9D50B7yH5A';
GRANT ALL PRIVILEGES ON DATABASE tucanbit TO tucanbit;
ALTER USER tucanbit CREATEDB;
\q
EOF

# Configure PostgreSQL for remote connections
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/*/main/postgresql.conf
echo "host all all 0.0.0.0/0 md5" | sudo tee -a /etc/postgresql/*/main/pg_hba.conf

# Restart PostgreSQL
sudo systemctl restart postgresql
sudo systemctl status postgresql
