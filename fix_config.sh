# Fix for TucanBIT config file issue
# The application is looking for 'config' but the file is 'config.yaml'

# Option 1: Create a symlink (recommended)
ln -sf config.yaml config/config

# Option 2: Set environment variable
export CONFIG_NAME=config.yaml

# Option 3: Copy the file with the expected name
cp config/config.yaml config/config

echo 'Config file issue fixed!'
echo 'You can now run: ./tucanbit'
