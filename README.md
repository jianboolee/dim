# 关于本项目

这是一个轻量的 IM 项目，目标是一个 web 版本优先、适配移动端使用的项目，包括简单的消息发送与接收。

## 模块

- **im-backend**：Go + Gin + MongoDB，REST `/im/api`、WebSocket `/im/ws`
- **im-frontend**：Vue 3 + Vite + Pinia

## 集成方式

IM **不提供登录注册**。业务系统通过服务端接口创建会话并跳转用户进入聊天。

详见 [docs/INTEGRATION.md](./docs/INTEGRATION.md)。

## 快速开始

```bash
# 后端
cd im-backend
cp .env.example .env   # 编辑 JWT_SECRET、INTEGRATION_API_KEY 等
go run ./cmd/migrate
go run ./cmd/server

# 前端
cd im-frontend
npm install
npm run dev
```

开发环境下 Vite 将 `/im/api`、`/im/ws` 分别代理到 `localhost:8901`、`localhost:8902`（API 与 WS 分进程部署时）。

