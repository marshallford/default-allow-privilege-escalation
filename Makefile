default: test

VERSION := 1.0.2-dev

GITHUB_REPOSITORY ?= marshallford/default-allow-privilege-escalation
IMAGE := $(GITHUB_REPOSITORY)

GIT_BRANCH ?= `git rev-parse --abbrev-ref HEAD`
GIT_COMMIT ?= `git rev-parse --short HEAD`

ifeq (`tty > /dev/null && echo 1 || echo 0`, 1)
DOCKER_FLAGS := --rm -it
else
DOCKER_FLAGS := --rm
endif
DOCKER := docker

GO := go
GOARCH := `go env GOARCH`
GOOS := `go env GOOS`
BIN := default-allow-privilege-escalation-$(GOOS)_$(GOARCH)

KUSTOMIZE := kustomize

lint:
	$(DOCKER) run --pull --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.29 golangci-lint run -v

test:
	$(GO) test ./... -race

coverage:
	$(GO) test ./... -race -coverpkg=./... -covermode=atomic $(if $(CI), -coverprofile=coverage.out)

build:
	$(GO) build -o $(BIN) cmd/webhook/main.go

docker-build:
	$(DOCKER) build --pull . \
		--build-arg MAINTAINER="Marshall Ford <inbox@marshallford.me>" \
		--build-arg CREATED=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		--build-arg REVISION=$(GIT_COMMIT) \
		--build-arg VERSION=$(VERSION) \
		--build-arg TITLE=$(IMAGE) \
		--build-arg REPOSITORY_URL=https://github.com/$(GITHUB_REPOSITORY) \
		-t $(IMAGE):$(VERSION)

kubectl-install-build:
	$(KUSTOMIZE) build deploy > kubectl-install.yaml

docker-push:
	$(DOCKER) push $(IMAGE):$(VERSION)

run:
	$(GO) run cmd/webhook/main.go

docker-run:
	$(DOCKER) run $(DOCKER_FLAGS) -p 8443:8443 $(IMAGE):$(VERSION)

.PHONY: lint test coverage build docker-build kubectl-install-build docker-push run docker-run
