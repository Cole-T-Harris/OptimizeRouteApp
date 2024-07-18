#!/bin/bash

# Build Go code
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o optimizeRoute optimizeRoute.go


# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi

# Package into ZIP file
zip -r lambda_function.zip optimizeRoute

if [ $? -ne 0 ]; then
    echo "ZIP creation failed. Exiting."
    exit 1
fi

# Optional: Display success message
echo "Go build completed. Output binary: optimizeRoute"