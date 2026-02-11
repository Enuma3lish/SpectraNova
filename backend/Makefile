.PHONY: api build run test

api:
	protoc --proto_path=./api \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		./api/fenzvideo/v1/*.proto

build:
	go build -o ./bin/fenzvideo ./cmd/fenzvideo

run:
	go run ./cmd/fenzvideo -conf ./configs/config.yaml

test:
	go test ./...
