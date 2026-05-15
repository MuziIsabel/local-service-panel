# TODO

> 本文件用于记录当前开发阶段的短期任务、当前状态、计划进展和阻塞事项。  
> `docs/ROADMAP.md` 负责版本路线图，`TODO.md` 负责近期可执行事项和日常进度跟踪。

## 当前状态

| 字段 | 内容 |
|---|---|
| 当前阶段 | v0.6 安装与打包 |
| 当前重点 | 无（v0.6 完成） |
| 最近完成 | v0.6 端到端验证通过 |
| 下一步 | 等待下阶段规划（v0.7） |
| 阻塞状态 | 无阻塞 |
| 最后更新 | 2026-05-15 |

## 进度总览

| 模块 | 状态 | 进度 |
|---|---|---|
| 项目基础 | Done | 27/27 |
| Go Agent 初始化 | Done | 9/9 |
| SQLite 初始化 | Done | 6/6 |
| Token 鉴权 | Done | 5/5 |
| Tauri + React UI 初始化 | Done | 8/8 |
| 开发体验 | Done | 5/5 |
| Windows Service 管理 | Done | 10/10 |
| 自定义软件管理 | Done | 10/10 |
| 自启管理 MVP | Done | 6/6 |
| 事件日志与仪表盘增强 | Done | 7/7 |

状态说明：

- `Planned`：已规划，尚未开始
- `Not Started`：本阶段内任务，尚未开始
- `In Progress`：正在进行
- `Blocked`：被阻塞
- `Done`：已完成

## 已完成阶段：v0.1 文档与项目骨架

### 1. 项目基础

- [x] 创建项目目录 `D:\agent\local-service-panel`
- [x] 创建 README.md
- [x] 创建 docs 文档目录
- [x] 创建 PRD 文档
- [x] 创建技术方案文档
- [x] 创建系统架构文档
- [x] 创建 MVP 范围文档
- [x] 创建数据模型文档
- [x] 创建 API 设计文档
- [x] 创建安全与权限设计文档
- [x] 创建开发路线图
- [x] 创建技术选型 ADR
- [x] 创建 TODO 文档
- [x] 创建 AGENTS.md 开发代理工作约定
- [x] 创建 Vibe Coding 协作指南
- [x] 创建 CONTEXT.md 项目上下文与领域语言
- [x] 创建开发工作流文档
- [x] 创建完成标准文档
- [x] 创建测试策略文档
- [x] 创建开发环境文档
- [x] 创建运行与排障手册
- [x] 创建 Agent Healthz 功能规格
- [x] 创建 Pi Prompt Template：plan-next
- [x] 创建 Pi Prompt Template：do-next
- [x] 创建 Pi Prompt Template：task
- [x] 创建 Pi Prompt Template：review-task
- [x] 创建 Pi Prompt Template：do-next-reviewed

### 2. Go Agent 初始化

- [x] 创建 `agent/` 目录
- [x] 初始化 Go module
- [x] 创建 Agent 基础目录结构
- [x] 实现配置加载模块
- [x] 实现结构化日志模块
- [x] 实现 `/api/healthz`
- [x] 实现 HTTP server 启动逻辑
- [x] 添加本地开发配置
- [x] 添加 Agent 启动脚本

### 3. SQLite 初始化

- [x] 选择 SQLite Go 驱动
- [x] 创建数据库连接模块
- [x] 创建 migrations 目录
- [x] 创建初始表结构
- [x] 实现数据库初始化逻辑
- [x] 添加基础 repository 示例

### 4. Token 鉴权

- [x] 设计 token 文件路径
- [x] 首次启动自动生成 token
- [x] 实现 Bearer token 中间件
- [x] 实现开发环境 token 配置
- [x] 避免 token 写入日志

### 5. Tauri + React UI 初始化

- [x] 创建 `app/` 目录
- [x] 初始化 Tauri 项目
- [x] 初始化 React + TypeScript
- [x] 安装 UI 组件库 Ant Design
- [x] 创建基础布局
- [x] 创建 Dashboard 空页面
- [x] 创建 API client
- [x] 调用 `/api/healthz` 显示 Agent 连接状态

### 6. 开发体验

- [x] 添加根目录开发脚本说明
- [x] 添加 Agent 开发启动说明
- [x] 添加 UI 开发启动说明
- [x] 补充 `.env.example`
- [x] 补充 CONTRIBUTING 草案，可选

## 已完成阶段：v0.2 Windows Service 管理

- [x] 创建 Windows Service 管理功能规格
- [x] 定义 Windows Service DTO 与 Provider 接口
- [x] 调研 Go 枚举 Windows Service 的实现方式
- [x] 实现服务列表 API
- [x] 实现服务详情 API
- [x] 实现启动服务 API
- [x] 实现停止服务 API
- [x] 实现重启服务 API
- [x] 实现修改启动类型 API
- [x] 前端创建服务列表页面
- [x] 前端添加服务操作按钮

## 已完成阶段：v0.3 自定义软件管理

- [x] 创建自定义软件管理功能规格
- [x] 完善 Custom App repository 和领域 DTO
- [x] 实现 Custom App 列表/详情 API
- [x] 实现添加 Custom App API
- [x] 实现编辑 Custom App API
- [x] 实现删除 Custom App API
- [x] 实现启动 Custom App 和 PID 记录
- [x] 实现停止 Custom App
- [x] 实现 stdout/stderr 日志写入和读取
- [x] 前端创建 Custom Apps 页面和添加/编辑表单

## 已完成阶段：v0.4 自启管理 MVP

- [x] 创建自启管理功能规格
- [x] 实现 autostart 包：命令构造、HKCU Run provider、非 Windows stub
- [x] 实现 Custom App autostart Service 方法
- [x] 实现 `POST /api/custom-apps/{id}/autostart`
- [x] 前端 Custom Apps 页面接入 AutoStart 开关
- [x] 补充 API 文档、测试和手动验证说明

## 当前阶段：v0.6 安装与打包

- [x] Agent Windows Service 改造：`-service install/uninstall/run` 标志
- [x] 安装脚本 `scripts/install.ps1`
- [x] 卸载脚本 `scripts/uninstall.ps1`
- [x] Agent Makefile：`make build` / `make install` / `make uninstall`
- [x] 版本号统一：0.6.0 同步到 agent/app/tauri
- [x] Tauri 打包配置：图标 + MSI 目标
- [x] Token 读取机制：Tauri Rust command + 前端适配
- [x] 功能规格：`docs/specs/installation-and-packaging.md`
- [x] 端到端验证：安装 → 启动 → 使用 → 卸载流程
- [x] 完善 README 安装使用说明

## 进展日志

### 2026-05-10

#### v0.1 里程碑完成

- 完成 Go Agent 初始化：目录结构、配置加载、结构化日志、HTTP server、`/api/healthz`。
- 完成 SQLite 初始化：`modernc.org/sqlite`、嵌入式 migrations、初始表、repository 示例。
- 完成 Token 鉴权：token 自动生成、Bearer middleware、healthz 白名单、开发环境 token。
- 完成 Tauri + React UI 初始化：Vite、React、TypeScript、Ant Design、Dashboard、healthz 轮询。
- 完成开发体验：README 快速开始、`.env.example`、CONTRIBUTING、开发命令说明。
- 完成项目治理文档：AGENTS、CONTEXT、WORKFLOW、DoD、TESTING、RUNBOOK、Prompt Templates。

#### v0.2 Windows Service 管理完成

- 完成 Windows Service 领域模型、DTO、Provider 接口、Windows SCM Provider 和非 Windows stub。
- 完成 Windows Service 列表、详情、启动、停止、重启、修改启动类型 API。
- 完成 protected service 保护策略。
- 完成前端 Services 页面、搜索过滤、状态标签、受保护标签和操作按钮。
- 前端 `npm run lint` 和 `npm run build` 验证通过。

#### v0.3 自定义软件管理完成

- 完成 Custom App 领域模型、DTO、repository 扩展和业务 Service。
- 完成 Custom App 列表、详情、创建、编辑、删除、启动、停止、日志 API。
- 完成进程启动、PID 记录、stdout/stderr 日志写入和读取。
- 完成前端 Custom Apps 页面、添加/编辑表单、启动/停止/删除/日志查看操作。
- 前端 `npm run lint` 和 `npm run build` 验证通过。

#### v0.4 自启管理完成

- 完成 autostart 包：命令构造、HKCU Run provider、非 Windows stub。
- 完成 Custom App autostart Service 方法和 `POST /api/custom-apps/{id}/autostart`。
- 完成前端 Custom Apps 页面 AutoStart 开关。
- 完成 API 文档、测试和手动验证说明。
- 已知限制：HKCU Run 无法表达 workingDir，复杂参数引用后续仍需强化。

#### v0.5 启动

- 同步 `docs/ROADMAP.md`，将 v0.4 标记为完成。
- 更新 `README.md`，切换当前阶段到 v0.5。
- 更新 `docs/specs/autostart-management.md`，增加 v0.4 实现状态、验收记录和已知限制。
- 更新 `docs/API.md`，补充 Custom App autostart API 和错误码。
- 创建 `docs/specs/event-log-and-dashboard.md`，定义事件日志、事件 API 和 Dashboard 增强规格。
- 更新 `TODO.md`，切换当前阶段到 v0.5 事件日志与仪表盘增强。

#### 实现事件日志 repository

- 创建 `internal/db/repository/event_log.go`：EventLogRepo 提供 Create 和 List（支持 targetId/targetType/action/status 过滤 + limit + DESC）。
- 创建 `internal/events/events.go`：Event 领域类型、DTO、Service（Record 不阻塞主流程、List）。
- 添加测试：Create+List、过滤、limit、DTO 映射。3 个测试全部 PASS。
- 更新 `TODO.md`。

#### 实现 GET /api/events

- 扩展 Handler：新增 `eventSvc *events.Service` 依赖。
- 实现 `GET /api/events` handler：支持 limit/targetId/targetType/action/status 过滤。
- 添加 `parseIntOrDefault` 工具函数。
- 更新 `main.go`：创建 EventLogRepo → EventService → 注入 Handler。
- 更新测试：`newTestHandler` 适配 3 参数签名。
- E2E 验证：端点正常返回。
- 更新 `TODO.md`。

#### 在 Custom App 操作中写入事件

- 为 6 个 Custom App handler 添加 eventSvc.Record 调用：
  - Create：成功 → CUSTOM_APP_CREATED，失败 → CUSTOM_APP_CREATE_FAILED
  - Update：成功 → CUSTOM_APP_UPDATED，失败 → CUSTOM_APP_UPDATE_FAILED
  - Delete：成功 → CUSTOM_APP_DELETED，失败 → CUSTOM_APP_DELETE_FAILED
  - Start：成功 → CUSTOM_APP_STARTED，失败 → CUSTOM_APP_START_FAILED
  - Stop：成功 → CUSTOM_APP_STOPPED，失败 → CUSTOM_APP_STOP_FAILED
  - Autostart：成功 → CUSTOM_APP_AUTOSTART_CHANGED，失败 → CUSTOM_APP_AUTOSTART_CHANGE_FAILED
#### 在 Windows Service 操作中写入事件

- 为 4 个 Windows Service handler 添加 eventSvc.Record 调用：
  - Start：成功 WINDOWS_SERVICE_STARTED，失败 WINDOWS_SERVICE_START_FAILED
  - Stop：成功 WINDOWS_SERVICE_STOPPED，失败 WINDOWS_SERVICE_STOP_FAILED
  - Restart：成功 WINDOWS_SERVICE_RESTARTED，失败 WINDOWS_SERVICE_RESTART_FAILED
  - SetStartType：成功 WINDOWS_SERVICE_START_TYPE_CHANGED，失败 WINDOWS_SERVICE_START_TYPE_CHANGE_FAILED
- 新增 TestMain 创建全局测试 events 服务，避免 nil pointer dereference。
- 所有 `go test ./...` PASS。
- 更新 `TODO.md`。

#### Dashboard 统计卡片与最近事件展示

- 前端新增 `EventLogDTO` 类型和 `listEvents()` API 客户端函数。
- 重写 Dashboard 页面，添加 4 张统计卡片：
  - Windows Services：总数 / 运行中数量
  - Custom Apps：总数 / 运行中数量
  - Agent Uptime：在线/离线状态
  - Recent Errors：最近 10 条事件中的错误数量
- Dashboard 增加最近事件表格（Table 组件），展示时间、Action、Status、Message。
- 失败事件以红色标签高亮。
- 使用 `Promise.allSettled` 并行获取 4 个 API，独立容错。
- 每 10 秒自动刷新。
- `npm run lint` 和 `npm run build` 通过。
- 更新 `TODO.md`。

#### v0.6 安装与打包

- **Agent Windows Service 改造**：新增 `agent/internal/service/service.go`，使用 `golang.org/x/sys/windows/svc` 实现服务生命周期。`main.go` 新增 `-service install/uninstall/run` 标志，抽取 `runAgent(ctx, dataDir)` 函数复用启动逻辑。
- **安装/卸载脚本**：创建 `scripts/install.ps1`（创建数据目录、复制二进制、注册 Service、启动）和 `scripts/uninstall.ps1`（停止、卸载、可选清理数据）。
- **版本信息统一**：创建 `agent/Makefile`；同步 version.go/package.json/tauri.conf.json/Cargo.toml 到 `0.6.0`；UI 侧边栏增加版本号显示。
- **Tauri 打包配置**：生成占位图标（32x32/128x128/icon.ico）；限制 bundle targets 为 `msi`。
- **Token 读取机制**：Tauri Rust 层新增 `read_token_file` 命令；前端 `getToken()` 增加 Tauri invoke 路径，回退到 `VITE_DEV_TOKEN`。
- **文档同步**：创建 `docs/specs/installation-and-packaging.md`；更新 ROADMAP/TODO/README/RUNBOOK/DEV_ENV。
- 所有 `go test ./...` PASS，`npm run lint && npm run build` PASS。

## TODO 更新约定

每完成或推进一个任务，都必须同步更新本文件。

更新规则：

1. 完成任务后，将对应任务从 `[ ]` 改为 `[x]`。
2. 如果一个模块开始执行，将“进度总览”中的状态改为 `In Progress`。
3. 如果一个模块完成，将状态改为 `Done`，并更新进度数字。
4. 如果遇到阻塞，将状态改为 `Blocked`，并在“阻塞事项”中说明原因。
5. 如果新增需求或问题，添加到“待确认问题”。
6. 每次明显推进后，在“进展日志”追加一条记录。
7. “当前状态”中的当前重点、最近完成、下一步、阻塞状态、最后更新需要保持同步。
8. 每完成一个阶段后，同步更新 `docs/ROADMAP.md`。

## 待确认问题

- [ ] Agent 默认端口是否固定为 `17645`
- [ ] 自定义软件 MVP 自启方式是否优先使用 HKCU Run
- [ ] UI 是否需要提供“启动 Agent/安装 Agent”的引导页面
- [ ] 是否优先使用 npm、pnpm 或 yarn
- [ ] v0.2 Windows Service 操作测试使用哪个安全测试服务
- [ ] v0.3 Custom App 测试程序使用 Go、Node 还是 Python
- [ ] v0.3 是否执行 stopCommand，还是仅保存到配置中
- [ ] v0.4 Custom App 自启是否接受 workingDir 无法通过 HKCU Run 表达的限制
- [ ] v0.5 是否新增独立 Events 页面，还是先只在 Dashboard 展示最近事件

## 阻塞事项

暂无。
