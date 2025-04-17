# note: call scripts from /scripts

.PHONY: proto build test clean

# Generate protobuf code
proto:
	cd proto && buf generate

# Build all services
build:
	@echo "Building Go services..."
	@for service in services/*; do \
		if [ -f $$service/go.mod ]; then \
			cd $$service && go build -o bin/$$(basename $$service) ./cmd/... && cd -; \
		fi \
	done

# Run tests
test:
	@echo "Running Go tests..."
	@for service in services/*; do \
		if [ -f $$service/go.mod ]; then \
			cd $$service && go test ./... && cd -; \
		fi \
	done

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@for service in services/*; do \
		if [ -d $$service/bin ]; then \
			rm -rf $$service/bin; \
		fi \
	done
	rm -rf proto/gen

# Run all services locally
run:
	docker compose up --build

# Install development tools
tools:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
