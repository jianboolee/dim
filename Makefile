SERVICES := api ws frontend

.PHONY: help dev dev-api dev-ws dev-frontend docker-build docker-push docker-release $(SERVICES:%=docker-build-%) $(SERVICES:%=docker-push-%) $(SERVICES:%=docker-release-%)

help:
	@echo "d-im root targets"
	@echo "  make dev                 Start API, WebSocket, and frontend dev servers"
	@echo "  make dev-api             Start API dev server"
	@echo "  make dev-ws              Start WebSocket dev server"
	@echo "  make dev-frontend        Start frontend dev server"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build        Build all service images"
	@echo "  make docker-push         Push all service images"
	@echo "  make docker-release      Build and push all service images"
	@echo ""
	@echo "Per service:"
	@echo "  make docker-release-api"
	@echo "  make docker-release-ws"
	@echo "  make docker-release-frontend"
	@echo ""
	@echo "Optional variables:"
	@echo "  IMAGE_TAG=<tag>"
	@echo "  DOCKER_PLATFORM=linux/amd64"

dev:
	+@$(MAKE) -j3 dev-api dev-ws dev-frontend

dev-api:
	+@$(MAKE) -C im-backend dev-api

dev-ws:
	+@$(MAKE) -C im-backend dev-ws

dev-frontend:
	+@$(MAKE) -C im-frontend dev

docker-build: $(SERVICES:%=docker-build-%)
docker-push: $(SERVICES:%=docker-push-%)
docker-release: $(SERVICES:%=docker-release-%)

docker-build-api:
	$(MAKE) -C im-backend/cmd/api-server docker-build

docker-push-api:
	$(MAKE) -C im-backend/cmd/api-server docker-push

docker-release-api:
	$(MAKE) -C im-backend/cmd/api-server docker-release

docker-build-ws:
	$(MAKE) -C im-backend/cmd/ws-server docker-build

docker-push-ws:
	$(MAKE) -C im-backend/cmd/ws-server docker-push

docker-release-ws:
	$(MAKE) -C im-backend/cmd/ws-server docker-release

docker-build-frontend:
	$(MAKE) -C im-frontend docker-build

docker-push-frontend:
	$(MAKE) -C im-frontend docker-push

docker-release-frontend:
	$(MAKE) -C im-frontend docker-release
