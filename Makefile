default: test

GIT_BRANCH ?= `git rev-parse --abbrev-ref HEAD`
GIT_COMMIT ?= `git rev-parse --short HEAD`

test:
	go test ./... -race

coverage:
	go test -coverpkg=./... ./... -race

build:
	go build -o bin/webhook cmd/webhook/main.go

run:
	go run cmd/webhook/main.go
