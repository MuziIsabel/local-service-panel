# AGENTS.md

本文件是 Pi/AI 编码代理在本项目中的工作约定。进入本项目开发时，应优先遵循本文件，并结合 `README.md`、`TODO.md` 与 `docs/` 下的设计文档执行。

## 1. 项目概述

项目名称：Local Service Panel

目标：构建一个个人电脑本地部署的服务管理面板，用于可视化管理 Windows 服务、开机自启项、自定义软件、后台进程、日志与健康状态。

当前技术栈：

```text
Agent: Go
Desktop Shell: Tauri
Frontend: React + TypeScript
UI: Ant Design
Database: SQLite
API: localhost HTTP，后续可升级为 Windows Named Pipe
Target OS: Windows first
```

总体架构：

```text
Tauri Desktop UI
  ↓ localhost API / IPC
Go Agent Windows Service
  ↓
Windows Service / Registry Run / Startup Folder / Task Scheduler / Custom App / Process
  ↓
SQLite + Logs + Config
```

## 2. 当前阶段

当前阶段：`v0.1 文档与项目骨架`

当前重点：

1. 初始化 `agent/` Go 后端项目。
2. 实现 Agent 基础 HTTP 服务。
3. 实现 `/api/healthz`。
4. 初始化 `app/` Tauri + React 前端项目。
5. 前端能连接 Agent 并显示连接状态。

## 3. 文档优先级

开发时按以下顺序理解项目要求：

1. `AGENTS.md`：AI/开发代理工作规则。
2. `TODO.md`：当前任务、进度、阻塞与下一步。
3. `CONTEXT.md`：项目上下文、领域语言和关键不变量。
4. `README.md`：项目入口和文档导航。
5. `docs/PRD.md`：产品需求。
6. `docs/TECHNICAL_DESIGN.md`：技术方案。
7. `docs/ARCHITECTURE.md`：系统架构。
8. `docs/MVP.md`：MVP 范围。
9. `docs/API.md`：API 约定。
10. `docs/DATA_MODEL.md`：数据模型。
11. `docs/SECURITY.md`：安全与权限约束。
12. `docs/VIBE_CODING.md`：AI 辅助开发与协作流程。
13. `docs/WORKFLOW.md`：标准开发工作流。
14. `docs/DEFINITION_OF_DONE.md`：任务完成标准。
15. `docs/TESTING.md`：测试策略。
16. `docs/DEV_ENV.md`：开发环境和常用命令。
17. `docs/RUNBOOK.md`：运行与排障手册。
18. `docs/specs/`：具体功能规格。
19. `docs/ROADMAP.md`：版本路线图。
20. `docs/adr/`：架构决策记录。

如果代码实现与文档冲突，应先指出冲突，并优先更新或确认文档，不要直接无视文档继续实现。

## 4. TODO 维护规则

`TODO.md` 是本项目的轻量项目管理面板。每次完成、推进或阻塞任务，都必须同步更新。

必须遵守：

1. 完成任务后，将对应任务从 `[ ]` 改为 `[x]`。
2. 模块开始执行后，将“进度总览”中的状态改为 `In Progress`。
3. 模块完成后，将状态改为 `Done`，并更新进度数字。
4. 遇到阻塞时，将状态改为 `Blocked`，并在“阻塞事项”中说明原因。
5. 新增需求、疑问或决策点，应添加到“待确认问题”。
6. 每次明显推进后，应在“进展日志”追加记录。
7. “当前状态”中的当前重点、最近完成、下一步、阻塞状态、最后更新应保持同步。
8. 每完成一个阶段后，同步更新 `docs/ROADMAP.md`。

除非用户明确要求不更新，否则代码或文档有实际推进时，都应更新 `TODO.md`。

## 5. 文档同步规则

修改实现时，如果影响以下内容，必须同步更新对应文档：

| 变化类型 | 需要更新 |
|---|---|
| 产品功能变化 | `docs/PRD.md`, `docs/MVP.md` |
| 技术栈变化 | `docs/TECHNICAL_DESIGN.md`, `docs/adr/` |
| 架构变化 | `docs/ARCHITECTURE.md`, `CONTEXT.md`, `docs/adr/` |
| API 变化 | `docs/API.md`, 必要时更新 `docs/specs/` |
| 数据库表或领域模型变化 | `docs/DATA_MODEL.md`, `CONTEXT.md` |
| 权限、安全、认证变化 | `docs/SECURITY.md`, `docs/RUNBOOK.md` |
| 开发命令或环境变化 | `docs/DEV_ENV.md`, `README.md` |
| 测试策略变化 | `docs/TESTING.md` |
| 工作流程变化 | `docs/WORKFLOW.md`, `docs/VIBE_CODING.md`, `AGENTS.md` |
| 阶段计划变化 | `TODO.md`, `docs/ROADMAP.md` |

重大技术选择需要新增 ADR：

```text
docs/adr/0002-some-decision.md
```

## 6. 开发原则

### 6.1 先 MVP，后扩展

优先完成 MVP：

- Windows only
- Go Agent
- Tauri + React UI
- SQLite
- Windows Service 管理
- 自定义软件管理
- 基础开机自启管理

暂不做：

- 云同步
- 多用户权限系统
- 远程公网访问
- 插件市场
- 全平台同时支持
- Kubernetes 管理
- 复杂告警系统

### 6.2 Agent 和 UI 职责分离

Agent 负责：

- 系统服务管理
- 进程管理
- 自启管理
- SQLite 读写
- 日志与事件
- 权限敏感操作

UI 负责：

- 展示状态
- 表单输入
- 调用 Agent API
- 用户确认

UI 不应直接执行管理员操作，不应直接修改注册表，不应绕过 Agent。

### 6.3 本地优先

默认只服务本机：

```text
127.0.0.1
```

不要默认监听：

```text
0.0.0.0
公网 IP
局域网 IP
```

## 7. 安全约束

必须遵守：

1. Agent API 默认只监听 `127.0.0.1`。
2. 非 healthz API 需要 Bearer token。
3. token 不得写入日志。
4. 不要使用 `Access-Control-Allow-Origin: *` 作为默认策略。
5. 执行自定义软件时，优先使用 executable path + args array，不要默认使用 shell 字符串。
6. 高危操作需要前端确认，例如停止关键服务、禁用服务、强制杀进程。
7. 日志中应避免输出密码、token、secret、api key。
8. 不能默认开放远程管理能力。

涉及安全设计变化时，必须同步更新 `docs/SECURITY.md`。

## 8. Go Agent 开发约定

计划目录：

```text
agent/
  cmd/agent/
  internal/api/
  internal/auth/
  internal/config/
  internal/db/
  internal/domain/
  internal/service/
  internal/startup/
  internal/process/
  internal/customapp/
  internal/health/
  internal/logging/
  internal/events/
```

约定：

- 使用 Go module。
- 内部包放在 `internal/`。
- HTTP API 层不直接写业务细节，应调用 service/usecase 层。
- Windows 系统调用应尽量隔离，方便后续跨平台。
- API 错误响应应遵循 `docs/API.md`。
- 数据库 schema 变化应走 migrations。
- 不要把本地绝对路径硬编码在业务逻辑中，应放配置。

## 9. 前端开发约定

计划目录：

```text
app/
  src/
    api/
    components/
    pages/
    routes/
    store/
    types/
```

约定：

- 使用 React + TypeScript。
- UI 组件库使用 Ant Design。
- API 类型应尽量与 `docs/API.md` 对齐。
- 页面先做清晰可用，不追求复杂动效。
- 危险操作按钮必须有确认。
- Agent 不可用时，应展示明确连接状态和下一步提示。

## 10. 数据与配置约定

正式环境建议目录：

```text
C:\ProgramData\LocalServicePanel\
  config\
  data\
  logs\
```

开发环境可使用：

```text
.data/
  config/
  data/
  logs/
```

数据库：SQLite。

数据库结构应参考：

```text
docs/DATA_MODEL.md
```

## 11. API 约定

MVP API 以本机 HTTP 为主：

```text
http://127.0.0.1:17645
```

基础接口：

```text
GET /api/healthz
GET /api/targets
POST /api/targets/{id}/start
POST /api/targets/{id}/stop
POST /api/targets/{id}/restart
POST /api/targets/{id}/autostart
```

新增或修改 API 时，必须同步更新 `docs/API.md`。

## 12. 测试与验证约定

修改代码后应尽量执行相应验证。

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

具体命令以后以实际 `package.json`、`go.mod` 和脚本为准。命令变化时应更新 README 或开发说明。

## 13. 工作流

推荐每次开发遵循：

1. 查看 `TODO.md` 当前状态。
2. 必要时阅读 `CONTEXT.md`，确认领域语言和关键不变量。
3. 确认本次要做的任务。
4. 必要时阅读对应 docs，包括 `docs/VIBE_CODING.md` 与 `docs/WORKFLOW.md` 中的协作流程。
5. 对具体功能优先查看或创建 `docs/specs/` 中的功能规格。
6. 实现最小可用改动。
7. 运行可行的验证命令。
8. 按 `docs/DEFINITION_OF_DONE.md` 检查是否完成。
9. 更新 `TODO.md`。
10. 如有接口、数据模型、架构、安全变化，同步更新 docs。
11. 简要汇报修改内容、验证结果和下一步建议。

## 14. Pi Prompt Templates

本项目提供项目级 Pi 提示词模板：

```text
.pi/prompts/plan-next.md
.pi/prompts/do-next.md
.pi/prompts/do-next-reviewed.md
.pi/prompts/task.md
.pi/prompts/review-task.md
```

常用命令：

```text
/plan-next          # 规划 TODO 下一项任务，但不写代码
/do-next            # 执行 TODO 下一项任务，并更新 TODO
/do-next-reviewed   # 执行 TODO 下一项任务，完成后自检并修复问题
/task <任务描述>    # 按项目流程执行指定任务
/review-task        # 按项目约定审查当前改动
```

新增或修改模板后，如果 Pi 已经启动，需要使用 `/reload` 重新加载。

## 15. 不要做的事

除非用户明确要求，不要：

- 一次性引入过多框架。
- 跳过 MVP 直接做复杂功能。
- 默认开放远程访问。
- 把 token、密钥、密码写进代码或日志。
- 在未说明风险时修改系统服务关键逻辑。
- 直接删除用户文件。
- 忽略 `TODO.md` 进度更新。

## 16. 当前默认决策

- Agent 默认端口暂定为 `17645`。
- 自定义软件 MVP 自启方式优先考虑 `HKCU Run`。
- Windows Service 自启通过修改服务启动类型实现。
- UI 通过本机 API 调用 Agent。
- 后续安全增强可考虑 Windows Named Pipe。
