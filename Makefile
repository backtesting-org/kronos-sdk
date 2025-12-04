.PHONY: generate-proto
generate-proto:
	@echo "Generating gRPC code from proto files..."
	@mkdir -p grpc/gen/inference
	protoc --go_out=. --go_opt=module=github.com/backtesting-org/kronos-sdk \
		--go-grpc_out=. --go-grpc_opt=module=github.com/backtesting-org/kronos-sdk \
		grpc/proto/inference.proto
	@echo "Generated gRPC code successfully"

.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	mockery
	@echo "Generated mocks successfully"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  generate-proto  - Generate Go code from proto files"
	@echo "  generate-mocks  - Generate mocks from interfaces"
