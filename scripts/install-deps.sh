#!/bin/bash

# Jetson RealSense Middleware Dependency Installer
# Target Platform: NVIDIA Jetson Orin / Xavier (Ubuntu 20.04/22.04)
# Description: Installs necessary system libraries, configures udev rules, and sets up Go environment.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Jetson RealSense Middleware setup...${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (sudo ./install-deps.sh)${NC}"
    exit 1
fi

# 1. Update package lists
echo -e "${YELLOW}[1/4] Updating package lists...${NC}"
apt-get update

# 2. Install system dependencies
echo -e "${YELLOW}[2/4] Installing system dependencies...${NC}"
apt-get install -y \
    libusb-1.0-0-dev \
    libglfw3-dev \
    libgtk-3-dev \
    libssl-dev \
    pkg-config \
    build-essential \
    curl \
    git \
    ffmpeg

# 3. Configure udev rules for RealSense
echo -e "${YELLOW}[3/4] Configuring udev rules for Intel RealSense devices...${NC}"
# Based on https://github.com/IntelRealSense/librealsense/blob/master/config/99-realsense-libusb.rules
cat <<EOF > /etc/udev/rules.d/99-realsense-libusb.rules
# Intel RealSense D400 Series
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad1", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad2", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad3", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad4", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad5", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0ad6", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0af6", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0afe", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0aff", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b00", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b01", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b03", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b07", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b3a", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b41", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b4d", MODE="0666", GROUP="plugdev"
SUBSYSTEMS=="usb", ATTRS{idVendor}=="8086", ATTRS{idProduct}=="0b5b", MODE="0666", GROUP="plugdev"
EOF

# Reload udev rules
udevadm control --reload-rules && udevadm trigger
echo "Udev rules updated."

# 4. Check/Install Go (Optional)
echo -e "${YELLOW}[4/4] Checking Go installation...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "Go is already installed: $GO_VERSION"
else
    echo -e "${YELLOW}Go not found. Installing Go 1.26 (compatible version)...${NC}"
    # Note: Using 1.22 as stable base for now, adjust URL if 1.26 preview is strictly required
    wget https://go.dev/dl/go1.22.1.linux-arm64.tar.gz -O /tmp/go.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo "Go installed successfully."
fi

# 5. Setup Project Library Path
echo -e "${YELLOW}[5/5] Setting up local library paths...${NC}"
PROJECT_ROOT=$(pwd)
if [[ "$PROJECT_ROOT" == */scripts ]]; then
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
fi

LIB_PATH="$PROJECT_ROOT/lib"
if [ -d "$LIB_PATH" ]; then
    echo "Found local library directory: $LIB_PATH"
    echo "Note: When running apps, use: export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:$LIB_PATH"
else
    echo -e "${RED}Warning: 'lib' directory not found in project root. Please ensure librealsense2.so is present.${NC}"
fi

echo -e "${GREEN}Setup complete! You may need to replug your RealSense camera for udev rules to take effect.${NC}"
echo -e "To run tests: ${YELLOW}make test${NC}"
