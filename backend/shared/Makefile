PROTO_BASE_PATH = api/proto
OUT_BASE_PATH = pkg/pb

service ?= *

.PHONY: proto
proto:
	@echo "Generating code for service: $(service)..."
	@mkdir -p $(OUT_BASE_PATH)
	
	protoc -I $(PROTO_BASE_PATH) \
		--go_out=$(OUT_BASE_PATH) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_BASE_PATH) --go-grpc_opt=paths=source_relative \
		$(shell find $(PROTO_BASE_PATH)/$(service) -name "*.proto")

	@echo "Success! Files generated in $(OUT_BASE_PATH)"

.PHONY: proto-clean
proto-clean:
	@echo "Cleaning generated code..."
	find $(OUT_BASE_PATH) -name "*.pb.go" -delete