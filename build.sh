#!/bin/bash

# Set environment variables for Linux and AMD64 architecture
GOOS=linux
GOARCH=amd64

# Build Go code
go build -o optimizeRoute optimizeRoute.go

# Package into ZIP file
zip -r lambda_function.zip optimizeRoute

# Optional: Display success message
echo "Go build completed. Output binary: optimizeRoute"