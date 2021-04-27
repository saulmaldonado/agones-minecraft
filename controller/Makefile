BUILD_PATH := build

USERNAME := saulmaldonado
IMAGE_NAME := agones-mc-dns-controller
FULL_IMAGE_NAME := $(USERNAME)/$(IMAGE_NAME)
CONTAINER_NAME := $(IMAGE_NAME)

.PHONY: build

build: build-container tag-latest

build-container:
	docker build -t $(FULL_IMAGE_NAME):$(TAG) .

tag-latest:
	docker tag $(FULL_IMAGE_NAME):$(TAG) $(FULL_IMAGE_NAME):latest

run:
	docker run -it --rm --name $(CONTAINER_NAME) \
	-v $(HOME)/.kube/config:/root/.kube/config \
	-v $(HOME)/.config/gcloud/:/root/.config/gcloud/ \
	-e HOME=/root \
	$(FULL_IMAGE_NAME):$(TAG) \
	--gcp-project=$(GCP_PROJECT) \
	--zone=$(MANAGED_ZONE)

publish:
	docker push $(FULL_IMAGE_NAME) --all-tags

stop:
	-docker container stop $(shell docker container ls -q --filter name=$(CONTAINER_NAME)*)

delete-containers:
	-docker rm $(shell docker ps -a -q --filter name=$(CONTAINER_NAME)*)

delete-images:
	-docker rmi $(shell docker images -q $(FULL_IMAGE_NAME)) -f

clean: stop delete-containers delete-images

go-build:
	go build -o $(BUILD_PATH)/controller main.go

go-run:
	go run main.go
