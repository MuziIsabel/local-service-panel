# MVP 范围

## 1. MVP 目标

在 Windows 上实现一个可运行的本机服务管理面板，能够：

- 展示 Windows Services。
- 管理 Windows Services 的启动、停止、重启和启动类型。
- 添加自定义软件。
- 启动、停止自定义软件。
- 配置自定义软件开机自启。
- 通过桌面 UI 操作。
- 使用 SQLite 持久化配置。

## 2. MVP 必做功能

### 2.1 Agent 基础

- Go 项目初始化。
- 配置文件加载。
- SQLite 初始化。
- HTTP API 服务。
- token 鉴权。
- 结构化日志。
- Windows Service 安装/卸载脚本，允许后续实现。

### 2.2 Windows Service 管理

- 枚举服务列表。
- 查询服务状态。
- 查询服务启动类型。
- 启动服务。
- 停止服务。
- 重启服务。
- 修改启动类型：自动、手动、禁用。

### 2.3 自定义软件管理

- 添加自定义软件：名称、路径、工作目录、参数。
- 编辑自定义软件。
- 删除自定义软件。
- 启动自定义软件。
- 停止自定义软件。
- 记录 PID。
- stdout/stderr 写入日志文件。
- 设置是否开机自启。

### 2.4 UI 页面

- Dashboard 页面。
- 服务列表页面。
- 自定义软件页面。
- 添加/编辑自定义软件弹窗或页面。
- 设置页面。

### 2.5 数据持久化

- 保存自定义软件配置。
- 保存自启设置。
- 保存事件日志。
- 保存基础系统设置。

## 3. MVP 不做功能

- Docker 管理。
- Linux/macOS 支持。
- 远程管理。
- 云同步。
- 多用户权限。
- 插件市场。
- 复杂告警系统。
- 完整计划任务管理。
- 完整注册表启动项扫描。
- WinSW 集成。

## 4. 里程碑

### Milestone 1：文档与项目骨架

- 创建项目目录。
- 创建 docs 文档。
- 初始化 Go Agent。
- 初始化 Tauri + React App。

验收：项目可以启动空白 UI 和空白 Agent health check。

### Milestone 2：Agent API 与数据库

- SQLite 初始化。
- token 鉴权。
- `/api/healthz`。
- `/api/settings`。
- 统一错误响应。

验收：UI 能连接 Agent 并显示连接状态。

### Milestone 3：Windows Service 管理

- 枚举服务。
- 启停服务。
- 重启服务。
- 修改启动类型。

验收：可以从 UI 控制一个测试服务。

### Milestone 4：自定义软件管理

- 添加自定义软件。
- 启动/停止。
- 日志输出。
- PID 状态刷新。

验收：可以添加一个 exe 或脚本并从 UI 控制。

### Milestone 5：自启管理 MVP

- 自定义软件 HKCU Run 自启。
- Windows Service 自动/手动/禁用。

验收：重启电脑后配置生效。

### Milestone 6：打包与安装

- Agent 安装脚本。
- UI 打包。
- 基础 README 使用说明。

验收：可以在一台 Windows 机器上安装并使用。

## 5. 验收清单

- [ ] Agent 可启动。
- [ ] UI 可打开。
- [ ] UI 可连接 Agent。
- [ ] 可以看到 Windows 服务列表。
- [ ] 可以启动 Windows 服务。
- [ ] 可以停止 Windows 服务。
- [ ] 可以重启 Windows 服务。
- [ ] 可以修改 Windows 服务启动类型。
- [ ] 可以添加自定义软件。
- [ ] 可以启动自定义软件。
- [ ] 可以停止自定义软件。
- [ ] 可以设置自定义软件开机自启。
- [ ] 配置可持久保存。
