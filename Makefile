# proto 코드 생성 경로
PROTO_GEN_PATH := ./gen
PROTO_GEN_GO := $(PROTO_GEN_PATH)/go
PROTO_GEN_PYTHON := $(PROTO_GEN_PATH)/python

# proto 옵션
PROTO_OPT_PROTO_PATH := "--proto_path=./proto"
PROTO_OPT_GO_OUT := "--go_out=$(PROTO_GEN_GO)"
PROTO_OPT_GO_PATH := "--go_opt=paths=source_relative"
PROTO_OPT_GO_GRPC_OUT := "--go-grpc_out=$(PROTO_GEN_GO)"
PROTO_OPT_GO_GRPC_PATH := "--go-grpc_opt=paths=source_relative"
PROTO_OPT_PYTHON_OUT := "--python_out=$(PROTO_GEN_PYTHON)"
PROTO_OPT_PYTHON_GRPC_OUT := "--grpc_python_out=$(PROTO_GEN_PYTHON)"
PROTO_OPT_PYI_OUT := "--pyi_out=$(PROTO_GEN_PYTHON)"

# proto 타겟
PROTO_TARGETS := \
	retrieval/issue/v1 \
	retrieval/passage/v1

.PHONY: proto
proto:
	@mkdir -p $(PROTO_GEN_GO)
	@mkdir -p $(PROTO_GEN_PYTHON)
	@for target in $(PROTO_TARGETS); do \
		echo generating $$target; \
		for file in $$(ls ./proto/$$target); do \
			protoc $(PROTO_OPT_PROTO_PATH) $(PROTO_OPT_GO_OUT) $(PROTO_OPT_GO_PATH) $(PROTO_OPT_GO_GRPC_OUT) $(PROTO_OPT_GO_GRPC_PATH) "./proto/$$target/$$file"; \
			python3 -m grpc_tools.protoc $(PROTO_OPT_PROTO_PATH) $(PROTO_OPT_PYTHON_OUT) $(PROTO_OPT_PYTHON_GRPC_OUT) $(PROTO_OPT_PYI_OUT) "./proto/$$target/$$file"; \
		done; \
	done

.PHONY: prototype
prototype:
	go build -o ./bin/prototype ./cmd/prototype

.PHONY: collector
collector:
	go build -o ./bin/collector ./cmd/collector

.PHONY: clean
clean:
	@rm -rf ./bin
