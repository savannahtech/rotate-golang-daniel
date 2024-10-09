#!/bin/bash
set -e

DIRECTORY=${1:-"$HOME/Downloads/"}

# Ensure the directory path ends with a trailing slash
if [[ "${DIRECTORY}" != */ ]]; then
  DIRECTORY="${DIRECTORY}/"
fi

# Check if osquery is installed using brew
if ! brew list --cask | grep -q "osquery" && [ ! -f "/opt/osquery/lib/osquery.app/Contents/MacOS/osqueryd" ]; then
    echo "osquery not found and osqueryd binary is missing. Installing osquery via Homebrew..."
    brew install --cask osquery
    echo "osquery installed successfully."
else
    echo "osquery is already installed or the osqueryd binary exists."
fi

GO_BIN=$(which go)
if ! command -v wails &> /dev/null
then
    echo "'wails' could not be found, installing it now..."
    $GO_BIN install github.com/wailsapp/wails/v2/cmd/wails@latest

    if [ $? -eq 0 ]; then
        echo "'wails' installed successfully."
    else
        echo "Failed to install 'wails'. Please check your Go environment."
        exit 1
    fi
else
    echo "'wails' is already installed."
fi


JSON_CONTENT="{\"file_paths\": {\"downloads\": [\"${DIRECTORY}%%\"]}}"

echo $JSON_CONTENT | sudo tee /var/osquery/osquery.conf > /dev/null

echo "Configuration added to /var/osquery/osquery.conf"


# YAML content to be written to config.yaml
YAML_CONTENT=$(cat <<EOF
directory: '${DIRECTORY}'
check_frequency: 1 # in seconds
reporting_api: 'http://api.external.com/v1/report'
http_port: '9000'
socket_path: '/var/osquery/osquery.em'
mongo_uri: 'mongodb://user:password@localhost:27017'
EOF
)


CONFIG_FILE="config.yaml"

echo "$YAML_CONTENT" | sudo tee "$CONFIG_FILE" > /dev/null

echo "YAML configuration added to $CONFIG_FILE"

echo "setup complete!!"