#!/bin/bash

# Variables
INSTALL_DIR="/usr/share/man/man1"

# Check if the current directory contains the expected structure
if [ ! -d "./usr/share/man/man1" ]; then
    echo "Error: The expected directory structure './usr/share/man/man1' is not found."
    exit 1
fi

# Find and install man pages
echo "Installing man pages..."
find ./usr/share/man/man1 -type f -name "*.1" -exec install -v -m 644 {} "$INSTALL_DIR" \;

echo "Man pages have been installed to $INSTALL_DIR."
