SHELL=/bin/bash -o pipefail

TAG := $(strip $(shell git describe --tags --exact-match HEAD 2> /dev/null | cut -c 2- || git rev-parse --short HEAD))

CONTROLLER_PATH := controller
TERRAFORM_PATH := terraform
USERNAME := saulmaldonado

include .env
export

.PHONY: build terraform

all:
	@$(MAKE) -C $(CONTROLLER_PATH)

build:
	@$(MAKE) -C $(CONTROLLER_PATH) build

publish:
	@$(MAKE) -C $(CONTROLLER_PATH) publish

controller-build:
	@$(MAKE) -C $(CONTROLLER_PATH) build

controller-clean:
	@$(MAKE) -C $(CONTROLLER_PATH) clean

controller-run:
	@$(MAKE) -C $(CONTROLLER_PATH) run

go-build:
	@$(MAKE) -C $(CONTROLLER_PATH) go-build

go-run:
	@$(MAKE) -C $(CONTROLLER_PATH) go-run

tf:
	@$(MAKE) -C $(TERRAFORM_PATH)
