DOCKER_REGISTRY = index.docker.io
IMAGE_NAME = vsync
IMAGE_VERSION = latest
IMAGE_ORG = flaccid
IMAGE_TAG = $(DOCKER_REGISTRY)/$(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION)
export DOCKER_BUILDKIT = 1

WORKING_DIR := $(shell pwd)

.DEFAULT_GOAL := docker-build

.PHONY: build push

go-deps:: ## fetch all the go dependencies
		go get ./...

go-build:: ## native golang build
		@make go-deps
		@go build -o bin/linux_amd64/vsync cmd/vsync/vsync.go

go-build-macos:: ## build macos64 binary
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin_amd64/vsync cmd/vsync/vsync.go

go-build-windows:: ## build windows64 binary
	go get -u github.com/konsorten/go-windows-terminal-sequences
	GOOS=windows GOARCH=amd64 go build -o bin/darwin_amd64/vsync cmd/vsync/vsync.go

docker-release:: docker-build docker-push ## builds and pushes the docker image to the registry

docker-push:: ## pushes the docker image to the registry
		@docker push $(IMAGE_TAG)

docker-build:: ## builds the docker image locally
		@echo http_proxy=$(HTTP_PROXY) http_proxy=$(HTTPS_PROXY)
		@echo building $(IMAGE_TAG)
		@docker build --pull \
			--build-arg=http_proxy=$(HTTP_PROXY) \
			--build-arg=https_proxy=$(HTTPS_PROXY) \
			-t $(IMAGE_TAG) $(WORKING_DIR)

docker-run:: ## runs the docker image locally
		@docker run \
			-it \
			$(DOCKER_REGISTRY)/$(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION)

helm-install:: ## installs using helm from chart in repo
		@helm install --name vsync ./charts/vsync

helm-upgrade:: ## upgrades deployed helm release
		@helm upgrade vsync ./charts/vsync

helm-purge:: ## deletes and purges deployed helm release
		@helm delete --purge vsync

helm-render:: ## prints out the rendered chart
		@helm install --dry-run --debug charts/vsync

helm-validate:: ## runs a lint on the helm chart
		@helm lint charts/vsync

install-ghr:: ## installs ghr
		@cd /tmp
		@wget https://github.com/tcnksm/ghr/releases/download/v0.12.2/ghr_v0.12.2_linux_amd64.tar.gz
		@tar zxvf ghr_v0.12.2_linux_amd64.tar.gz
		@sudo mv ghr_v0.12.2_linux_amd64/ghr /usr/local/bin/

vault-dev::	## runs a vault dev server in foreground
		@vault server -dev

# a help target including self-documenting targets (see the awk statement)
define HELP_TEXT
Usage: make [TARGET]... [MAKEVAR1=SOMETHING]...

Available targets:
endef
export HELP_TEXT
help: ## this help target
	@cat .banner
	@echo
	@echo "$$HELP_TEXT"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
		{printf "\033[36m%-30s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)
