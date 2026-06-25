IMAGE_REGISTRY ?= registry.cn-beijing.aliyuncs.com
IMAGE_NAMESPACE ?= jianboo
IMAGE_NAME ?= dim-api
BACKEND_DIR := ../..
IMAGE_TAG ?= $(shell git -C $(BACKEND_DIR) rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)
DOCKER_PLATFORM ?= linux/amd64
DOCKER_BUILDER ?= docker buildx

IMAGE := $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)
LATEST_IMAGE := $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):latest

.PHONY: help docker-build docker-push docker-release

help:
	@echo "dim-api docker targets"
	@echo "  make docker-build    Build linux/amd64 image locally"
	@echo "  make docker-push     Push IMAGE_TAG and latest"
	@echo "  make docker-release  Build then push"
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE_TAG=$(IMAGE_TAG)"
	@echo "  IMAGE=$(IMAGE)"

docker-build:
	$(DOCKER_BUILDER) build \
		--platform $(DOCKER_PLATFORM) \
		--target api-server-runtime \
		-t $(IMAGE) \
		-t $(LATEST_IMAGE) \
		--load \
		$(BACKEND_DIR)

docker-push:
	docker push $(IMAGE)
	docker push $(LATEST_IMAGE)

docker-release: docker-build docker-push
