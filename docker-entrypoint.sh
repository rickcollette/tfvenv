#!/bin/bash

# Source tfvenv activation script if it exists
if [ -f "$TFVENV_DIR/bin/activate.sh" ]; then
    echo "Activating tfvenv environment..."
    source "$TFVENV_DIR/bin/activate.sh"
fi

# Source .tvenvrc if it exists
if [ -f "$TFVENV_DIR/.tvenvrc" ]; then
    echo "Sourcing tfvenv configuration..."
    source "$TFVENV_DIR/.tvenvrc"
fi

# Execute the provided command or default to bash
exec "$@"
