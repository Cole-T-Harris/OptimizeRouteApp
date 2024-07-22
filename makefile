# Define variables
FUNCTIONS_DIRS := optimizeRoute commutesQueue
BUILD_DIR := dist
BINARY_NAMES := $(FUNCTIONS_DIRS)
ZIP_NAMES := $(addprefix $(BUILD_DIR)/, $(addsuffix .zip, $(BINARY_NAMES)))

# Create the build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Default target
all: $(BUILD_DIR) build

# Build and package each Go Lambda function
build: $(ZIP_NAMES)

# Build the Go function and create a zip file
$(BUILD_DIR)/%.zip: %
	@echo "Building $*..."
	(cd $* && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $* main.go)
	@echo "Setting executable permissions for $*..."
	@chmod +x $*/$*
	@echo "Moving binary to $(BUILD_DIR)/$*..."
	@mv $*/$* $(BUILD_DIR)/$*
	@echo "Packaging $(BUILD_DIR)/$*.zip..."
	@cd $(BUILD_DIR) && zip -r $*.zip $*

# Run tests for all functions
# test:
# 	for dir in $(FUNCTIONS_DIRS); do \
# 		(cd $$dir && go test ./...); \
# 	done

# Format code for all functions
fmt:
	for dir in $(FUNCTIONS_DIRS); do \
		(cd $$dir && gofmt -w .); \
	done

# Clean up build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Ensure dependencies are up-to-date for all functions
mod:
	for dir in $(FUNCTIONS_DIRS); do \
		(cd $$dir && go mod tidy); \
	done

# Print help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all       - Build all Go Lambda functions"
	@echo "  build     - Build each Go Lambda function"
	@echo "  fmt       - Format Go code for all functions"
	@echo "  clean     - Clean up build artifacts for all functions"
	@echo "  mod       - Ensure dependencies are up-to-date for all functions"
	@echo "  help      - Print this help message"
