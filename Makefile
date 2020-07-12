default: test

REPOSITORY ?= marshallford/default-allow-privilege-escalation
IMAGE := $(REPOSITORY):latest
GIT_BRANCH ?= `git rev-parse --abbrev-ref HEAD`
GIT_COMMIT ?= `git rev-parse --short HEAD`

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.28 golangci-lint run -v

test:
	go test ./... -race

coverage:
	go test -coverpkg=./... ./... -race

build:
	go build -o bin/webhook cmd/webhook/main.go

docker-build:
	docker build --pull . -t $(IMAGE)

run:
	go run cmd/webhook/main.go
