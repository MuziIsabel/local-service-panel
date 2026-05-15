# 开发路线图

## v0.1 文档与骨架

状态：Done

目标：建立项目基础。

- [x] 创建项目目录。
- [x] 创建初始文档。
- [x] 初始化 Go Agent。
- [x] 初始化 SQLite。
- [x] 初始化 Token 鉴权。
- [x] 初始化 Tauri + React UI。
- [x] 添加基础开发脚本。
- [x] Agent health check。
- [x] UI 连接 Agent。
- [x] 补充开发体验文档与 CONTRIBUTING。

## v0.2 Windows Service 管理

状态：Done

目标：可以可视化管理 Windows 服务。

- [x] 创建功能规格：`docs/specs/windows-service-management.md`。
- [x] 定义 Windows Service DTO 与 Provider 接口。
- [x] 枚举 Windows Services。
- [x] 查询服务状态。
- [x] 查询启动类型。
- [x] 查询服务详情。
- [x] 启动服务。
- [x] 停止服务。
- [x] 重启服务。
- [x] 修改启动类型。
- [x] 前端服务列表页面。
- [x] 搜索、过滤、状态标签。
- [x] 前端服务操作按钮。

## v0.3 自定义软件管理

状态：Done

目标：用户可以添加自己的软件并管理。

- [x] 创建功能规格：`docs/specs/custom-app-management.md`。
- [x] 完善 Custom App repository 和领域 DTO。
- [x] SQLite managed_targets 表。
- [x] 添加自定义软件 API。
- [x] 编辑自定义软件 API。
- [x] 删除自定义软件 API。
- [x] 启动自定义软件。
- [x] 停止自定义软件。
- [x] 记录 PID。
- [x] stdout/stderr 日志文件。
- [x] 日志读取 API。
- [x] 前端 Custom Apps 页面。
- [x] 前端添加/编辑表单。

## v0.4 自启管理 MVP

状态：Done

目标：支持基础开机自启。

- [x] 创建功能规格：`docs/specs/autostart-management.md`。
- [x] Windows Service 启动类型与自启开关联动。
- [x] 自定义软件 HKCU Run 自启。
- [x] 展示自启来源。
- [x] 自启启用/禁用操作。
- [x] 前端 Custom Apps 页面接入 AutoStart 开关。

## v0.5 事件日志与仪表盘增强

状态：Done

目标：记录关键操作事件，并增强 Dashboard 可观测性。

- [x] 创建功能规格：`docs/specs/event-log-and-dashboard.md`。
- [x] 实现事件日志 repository。
- [x] 实现 `GET /api/events`。
- [x] 在 Custom App 操作中写入事件。
- [x] 在 Windows Service 操作中写入事件。
- [x] Dashboard 统计卡片：服务数/自定义应用数/错误数。
- [x] Dashboard 展示最近事件和错误。

## v0.6 安装与打包

状态：Done

目标：可分发使用。

- [x] Agent Windows Service 改造：`-service install/uninstall/run` 标志。
- [x] 安装脚本 `scripts/install.ps1`。
- [x] 卸载脚本 `scripts/uninstall.ps1`。
- [x] Agent Makefile：`make build` / `make install` / `make uninstall`。
- [x] 版本号统一：0.6.0 同步到 agent/app/tauri。
- [x] Tauri 打包配置：图标 + MSI 目标。
- [x] Token 读取机制：Tauri Rust command + 前端适配。
- [x] 功能规格：`docs/specs/installation-and-packaging.md`。
- [x] 端到端验证：构建 → 前台启动 → API 测试 → 停止。
- [x] 完善 README 安装使用说明。

## 后续版本

### v0.7

- WinSW 集成。
- 一键将自定义软件注册为 Windows Service。
- 失败自动重启策略。

### v0.8

- Task Scheduler 扫描和管理。
- Startup Folder 扫描。
- Registry Run 完整扫描。

### v0.9

- Docker 容器管理。
- Compose 项目识别。

### v1.0

- 稳定版。
- 安装包。
- 完整文档。
- 常见问题诊断。
