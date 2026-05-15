# 开发工作流

## 1. 目标

本文件定义 Local Service Panel 的日常开发流程，尤其用于 AI 辅助开发和 vibe coding。目标是保证每轮开发都小步、可验证、可追踪、文档同步。

## 2. 标准开发循环

每轮开发遵循：

```text
阅读状态 → 明确目标 → 小步实现 → 验证 → 更新 TODO → 汇报
```

详细步骤：

1. 阅读 `AGENTS.md`、`TODO.md`。
2. 根据任务阅读相关文档，例如 `docs/API.md`、`docs/DATA_MODEL.md`、`docs/SECURITY.md`。
3. 明确本轮目标和非目标。
4. 只做最小必要改动。
5. 运行可行的验证命令。
6. 更新 `TODO.md`。
7. 如影响项目事实，同步更新 docs。
8. 汇报修改内容、验证结果、TODO 更新和下一步建议。

## 3. 本轮目标模板

每次开始实现前，建议先明确：

```text
本轮目标：
- ...

本轮非目标：
- ...

计划修改文件：
- ...

验收标准：
- ...

需要同步文档：
- ...
```

## 4. 任务粒度

推荐一次只做一个可验证目标。

好的任务：

- 初始化 Go module。
- 实现 `/api/healthz`。
- 添加 SQLite migration 初始化。
- 创建 Dashboard 空页面。
- 添加 API client 基础封装。

不好的任务：

- 一次性完成整个 Agent。
- 一次性完成所有 Windows Service 管理。
- 一次性完成安装器、权限、安全、UI 全链路。

## 5. 文档同步流程

如果本轮改动影响以下内容，必须同步文档：

| 改动 | 文档 |
|---|---|
| API | `docs/API.md` |
| 数据模型 | `docs/DATA_MODEL.md` |
| 架构边界 | `docs/ARCHITECTURE.md`, `CONTEXT.md` |
| 安全策略 | `docs/SECURITY.md` |
| 开发命令 | `docs/DEV_ENV.md`, `README.md` |
| 任务进度 | `TODO.md` |
| 版本路线 | `docs/ROADMAP.md` |
| 重大决策 | `docs/adr/` |

## 6. 验证流程

优先运行与本轮任务相关的最小验证。

Go Agent：

```bash
go test ./...
go run ./cmd/agent
```

前端：

```bash
npm run lint
npm run build
npm run dev
```

如果验证暂时无法运行，必须说明原因，例如：

- 当前尚未初始化 Go module。
- 当前尚未初始化 package.json。
- 当前环境缺少 Tauri/Rust 依赖。
- 当前操作需要 Windows 管理员权限。

## 7. 汇报格式

每轮结束建议使用：

```text
完成内容：
- ...

修改文件：
- ...

验证结果：
- ...

TODO 更新：
- ...

风险/注意：
- ...

下一步建议：
- ...
```

## 8. 处理不确定性

如果需求不清楚：

1. 先提出问题。
2. 给出推荐默认方案。
3. 不要直接实现高风险猜测。

可以直接实现的低风险默认决策：

- Agent 默认端口 `17645`。
- Agent 默认监听 `127.0.0.1`。
- MVP 自定义软件自启优先 HKCU Run。
- UI 与 Agent 通过 localhost HTTP 通信。

需要确认的高风险决策：

- 远程管理。
- 管理员权限安装器。
- 强制停止系统关键服务。
- 删除真实软件文件。
- 开放局域网访问。

## 9. 分支和提交建议

当前未强制 Git 流程。后续如果使用 Git，建议：

- 小步提交。
- 提交信息包含模块和目标。
- 文档和代码同步提交。
- 高风险改动单独提交。

提交示例：

```text
agent: add healthz endpoint
```

```text
docs: add workflow and testing strategy
```
