#!/bin/bash

# Install swag if not already installed
if ! command -v swag &> /dev/null; then
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate Swagger documentation
swag init -g cmd/main.go -o docs 