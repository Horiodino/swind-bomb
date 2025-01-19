#!/bin/bash

BINARY_NAME="swinds"
BUILD_DIR="bin"
SOURCE_FILE="main.go"
SERVICE_FILE="/etc/systemd/system/${BINARY_NAME}.service"

mkdir -p $BUILD_DIR

build() {
    echo "Building binary..."
    go build -o "${BUILD_DIR}/${BINARY_NAME}" $SOURCE_FILE
}

install() {
    echo "Installing binary..."
    build
    sudo cp "${BUILD_DIR}/${BINARY_NAME}" "/usr/local/bin/${BINARY_NAME}"
}

create_service() {
    echo "Creating systemd service..."
    install
    sudo bash -c "cat > $SERVICE_FILE << EOF
[Unit]
Description=Swinds Go Program
After=network.target

[Service]
ExecStart=/usr/local/bin/${BINARY_NAME}
WorkingDirectory=/usr/local/bin
Restart=always
User=root
Group=root
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=${BINARY_NAME}

[Install]
WantedBy=multi-user.target
EOF"
}

setup_service() {
    echo "Setting up service..."
    create_service
    sudo systemctl daemon-reload
    sudo systemctl start "${BINARY_NAME}.service"
    sudo systemctl enable "${BINARY_NAME}.service"
}

status() {
    echo "Checking service status..."
    sudo systemctl status "${BINARY_NAME}.service"
    echo -e "\nChecking service logs..."
    sudo journalctl -u "${BINARY_NAME}.service" -n 50 --no-pager
}

clean() {
    echo "Cleaning build directory..."
    rm -rf $BUILD_DIR
}

check() {
    echo "Checking if binary exists..."
    if [ -f "/usr/local/bin/${BINARY_NAME}" ]; then
        echo "Binary exists"
    else
        echo "Binary does not exist"
    fi
}

case "$1" in
    "build")
        build
        ;;
    "install")
        install
        ;;
    "create-service")
        create_service
        ;;
    "setup-service")
        setup_service
        ;;
    "status")
        status
        ;;
    "clean")
        clean
        ;;
    "check")
        check
        ;;
    *)
        echo "Usage: $0 {build|install|create-service|setup-service|status|clean|check}"
        exit 1
        ;;
esac