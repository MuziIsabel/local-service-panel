# Contributing

欢迎参与 Local Service Panel 的开发。

## 开发准备

参见 [README.md](README.md) 的「快速开始」和 [docs/DEV_ENV.md](docs/DEV_ENV.md)。

## 开发流程

1. 查看 [TODO.md](TODO.md) 当前状态。
2. 阅读 [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md) 了解完成标准。
3. 阅读 [docs/WORKFLOW.md](docs/WORKFLOW.md) 了解开发工作流。
4. 阅读 [docs/VIBE_CODING.md](docs/VIBE_CODING.md) 了解 AI 辅助开发协作方式。
5. 阅读 [docs/SECURITY.md](docs/SECURITY.md) 了解安全约束。

## 代码约定

- Go Agent 代码在 `agent/` 目录。
- 前端代码在 `app/` 目录。
- Agent API 设计参考 [docs/API.md](docs/API.md)。
- 数据模型参考 [docs/DATA_MODEL.md](docs/DATA_MODEL.md)。

## 验证

### Go Agent

```bash
cd agent
go mod tidy
go vet ./...
go test ./...
go run ./cmd/agent -data ../.data
```

### 前端

```bash
cd app
npm run lint
npm run build
npm run dev
```

## 文档同步

改动如果影响 API、数据模型、架构、安全策略或开发命令，必须同步更新对应文档（参见 [AGENTS.md](AGENTS.md) 第 5 节）。

## 提交信息

当前阶段无强制 Git 流程，推荐小步提交：

```
agent: add healthz endpoint
ui: add Dashboard page
docs: update DEV_ENV with auth section
```
