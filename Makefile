# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2021
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

export CGO_ENABLED ?= 0
DOCKER_CMD ?= $(shell which docker 2> /dev/null || which podman 2> /dev/null || echo docker)
DOCKER_COMPOSE_CMD ?= $(shell which docker-compose 2> /dev/null || echo docker-compose)
GO_CMD ?= $(shell which go 2> /dev/null || echo go)
IMAGE_VERSION ?= $(shell git describe --abbrev=0 --tags)
IMAGE_NAME=gwtester/pgw:$(IMAGE_VERSION)

test:
	$(GO_CMD) test -v ./...
run:
	$(GO_CMD) run cmd/main.go
.PHONY: build
build:
	sudo -E $(DOCKER_CMD) build -t $(IMAGE_NAME) .
	sudo -E $(DOCKER_CMD) image prune --force
push: test build
	docker-squash $(IMAGE_NAME)
	sudo -E $(DOCKER_CMD) push $(IMAGE_NAME)

.PHONY: lint
lint:
	sudo -E $(DOCKER_CMD) run --rm -v $$(pwd):/tmp/lint \
	-e RUN_LOCAL=true \
	-e LINTER_RULES_PATH=/ \
	-e VALIDATE_KUBERNETES_KUBEVAL=false \
	github/super-linter

deploy:
	sudo -E $(DOCKER_COMPOSE_CMD) --file deployments/docker-compose.yml \
	up --always-recreate-deps --detach
undeploy:
	sudo -E $(DOCKER_COMPOSE_CMD) --file deployments/docker-compose.yml \
	down --remove-orphans

logs:
	sudo -E $(DOCKER_COMPOSE_CMD) --file deployments/docker-compose.yml logs

system-test:
	@vagrant up --no-destroy-on-error
