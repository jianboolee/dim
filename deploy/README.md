# d-im Portainer Deployment

This directory contains examples for deploying three application services:

- `dim-frontend`: Vue static frontend served by Nginx
- `dim-api`: Go REST API service
- `dim-ws`: Go WebSocket service

MongoDB and Redis are expected to be existing external services.

## Build And Push Images

Login first:

```bash
docker login --username=jianboolee@163.com registry.cn-beijing.aliyuncs.com
```

Build and push the API image:

```bash
cd im-backend/cmd/api-server
make docker-release
```

Build and push the WebSocket image:

```bash
cd im-backend/cmd/ws-server
make docker-release
```

Build and push the frontend image:

```bash
cd im-frontend
make docker-release
```

All three Makefiles default to `DOCKER_PLATFORM=linux/amd64` and push two tags:

- `registry.cn-beijing.aliyuncs.com/jianboo/<image>:<git-short-sha>`
- `registry.cn-beijing.aliyuncs.com/jianboo/<image>:latest`

Override the tag when needed:

```bash
make docker-release IMAGE_TAG=2026-06-25
```

## Portainer

Use `portainer-stack.example.yml` as the Portainer stack template. For
production, copy `stack.env.min.example` to your deployment `.env`, fill every
value there, then load that `.env` in Portainer. `stack.env.example` contains
sample values for reference only. The stack template intentionally reads
configuration only from environment variables and does not define fallback
defaults, so the env file is the single source of truth.

The example exposes:

- frontend: host `8900` -> container `80`
- API: host `8901` -> container `8080`
- WebSocket: host `8902` -> container `9000`

## Server Nginx

Use `nginx.chat.d-im.com.example.conf` as the server Nginx reference for
`https://chat.d-im.com`.
