.DEFAULT_GOAL: build

GIT_VERSION := $(shell git describe --dirty --always --tags)
TAG?=$(GIT_VERSION)
IMAGE?=vault-kubernetes-dotenv-manager
PREFIX?=szeber/$(IMAGE)
ARCH?=amd64

build:
	go mod download
	CGO_ENABLED=0 go build -ldflags="-s -w" -o vault_kubernetes_dotenv_manager main.go

dockerBuild:
	@echo "version: $(PREFIX):$(TAG)"
	docker build -t $(PREFIX):$(TAG) .

dockerPush: dockerBuild
	docker push $(PREFIX):$(TAG)

dockerTagLatest: dockerPush
	docker tag  $(PREFIX):$(TAG)  $(PREFIX):latest
	docker push $(PREFIX):latest

clean:
	go clean -v .
	docker rmi $(PREFIX):$(TAG)

docker: dockerTagLatest

.PHONY: build