GOMOD := bursavich.dev/fieldmask
MAKEDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

protogen:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@protoc --go_out="$(MAKEDIR)" --go_opt=paths=import --go_opt=module=$(GOMOD) \
		--proto_path="$(MAKEDIR)" "$(MAKEDIR)"/internal/testpb/*.proto
