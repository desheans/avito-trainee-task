.PHONY: generate build test lint docker-up docker-down migrate

generate:
	oapi-codegen --config=./docs/config.yml -generate types,server ./docs/openapi.yml

test-all:
	go test -v ./inernal/tests/...

test-integration:
	go test -v ./internal/tests/integration/...

test-e2e:
	go test -v ./internal/tests/e2e/...

lint:
	golangci-lint run

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
