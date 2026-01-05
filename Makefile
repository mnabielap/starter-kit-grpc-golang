.PHONY: proto clean run build

# Ensure generated directory exists
init:
	mkdir -p api/gen/v1
	mkdir -p api/openapi/v1

# Generate code using buf (Recommended if buf is installed)
# If you don't have buf, we will use protoc directly below.
proto-buf:
	buf generate

# Standard protoc generation (Use this if you don't have buf CLI)
setup-googleapis:
	if [ ! -d "googleapis" ]; then \
		git clone https://github.com/googleapis/googleapis.git; \
	fi

proto: setup-googleapis
	protoc -I . \
		-I googleapis \
		--go_out ./api/gen --go_opt paths=source_relative \
		--go-grpc_out ./api/gen --go-grpc_opt paths=source_relative \
		--grpc-gateway_out ./api/gen --grpc-gateway_opt paths=source_relative \
		--openapiv2_out ./api/openapi --openapiv2_opt allow_merge=true,merge_file_name=apidocs \
		api/proto/v1/*.proto

run:
	go run cmd/server/main.go