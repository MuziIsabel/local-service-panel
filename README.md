# Local Service Panel

个人电脑本地服务管理面板。用于可视化管理 Windows 服务、开机自启项、自定义软件、后台进程与健康状态。

## 项目定位

本项目面向个人电脑场景，提供一个本机运行的管理面板：

- 查看所有服务/软件运行状态
- 管理 Windows Service 的启动、停止、重启
- 查看和修改开机自启状态
- 添加自定义软件进行统一管理
- 支持日志、健康检查、异常状态展示
- 后续扩展 Docker、Linux systemd、远程管理

## 推荐技术栈

当前采用：

```text
Agent: Go
Desktop Shell: Tauri
Frontend: React + TypeScript
UI: Ant Design
Database: SQLite
API: localhost HTTP，后续可切换 Named Pipe
Platform: Windows first
```

## 架构概览

```text
Tauri Desktop UI
  ↓ localhost API / IPC
Go Agent Windows Service
  ↓
Windows Service / Registry Run / Startup Folder / Task Scheduler / Custom App / Process
  ↓
SQLite + Logs + Config
```

## 文档入口

- [项目上下文与领域语言](CONTEXT.md)
- [AI/开发代理工作约定](AGENTS.md)
- [近期 TODO](TODO.md)
- [产品需求文档](docs/PRD.md)
- [技术方案](docs/TECHNICAL_DESIGN.md)
- [系统架构](docs/ARCHITECTURE.md)
- [MVP 范围](docs/MVP.md)
- [数据模型](docs/DATA_MODEL.md)
- [API 设计](docs/API.md)
- [安全与权限设计](docs/SECURITY.md)
- [Vibe Coding 协作指南](docs/VIBE_CODING.md)
- [开发工作流](docs/WORKFLOW.md)
- [完成标准](docs/DEFINITION_OF_DONE.md)
- [测试策略](docs/TESTING.md)
- [开发环境](docs/DEV_ENV.md)
- [运行与排障手册](docs/RUNBOOK.md)
- [开发路线图](docs/ROADMAP.md)
- [功能规格：Agent Healthz](docs/specs/agent-healthz.md)
- [功能规格：Windows Service 管理](docs/specs/windows-service-management.md)
- [功能规格：自定义软件管理](docs/specs/custom-app-management.md)
- [功能规格：自启管理 MVP](docs/specs/autostart-management.md)
- [功能规格：事件日志与仪表盘增强](docs/specs/event-log-and-dashboard.md)
- [功能规格：安装与打包](docs/specs/installation-and-packaging.md)
- [ADR 0001：技术栈选型](docs/adr/0001-tech-stack.md)
- [贡献指南](CONTRIBUTING.md)

## 目录规划

```text
local-service-panel/
  agent/                 # Go Agent
  app/                   # Tauri + React 前端
  docs/                  # 项目文档
  scripts/               # 开发脚本
  .data/                 # 开发环境本地数据（已 git ignore）
  README.md
```

## 快速开始

### 开发模式

```bash
# 启动 Agent（开发模式，前台进程）
cd agent
go run ./cmd/agent -data ../.data

# 或者使用启动脚本
bash scripts/agent-dev.sh

# 验证
curl http://127.0.0.1:17645/api/healthz
```

Agent 默认监听 `127.0.0.1:17645`，首次启动自动生成 token 并初始化 SQLite。

### 安装模式（Windows Service）

以**管理员身份**运行 PowerShell：

```powershell
# 1. 构建 Agent
cd agent
go build -ldflags "-X 'github.com/user/local-service-panel/agent/internal/version.Version=0.6.0'" -o agent.exe ./cmd/agent

# 2. 一键安装
.\scripts\install.ps1

# 或通过 Makefile
cd agent
make install
```

安装后：

- Agent 以 Windows Service `LocalServicePanelAgent` 运行。
- 数据目录：`C:\ProgramData\LocalServicePanel\`。
- 开机自启动。

### 卸载

```powershell
.\scripts\uninstall.ps1

# 如需清理数据
.\scripts\uninstall.ps1 -CleanData
```

### UI 打包

```bash
cd app
npm run lint
npm run build
npx tauri build        # 生成 MSI 安装包
```

生成的安装包位于 `app/src-tauri/target/release/bundle/msi/`。

### Agent 命令行参考

| 标志 | 说明 |
|---|---|
| `-version` | 打印版本号并退出 |
| `-data <path>` | 指定数据目录（默认 `.data/`） |
| `-service install` | 注册为 Windows Service |
| `-service uninstall` | 卸载 Windows Service |
| `-service run` | 以 Windows Service 模式运行 |

默认模式（无 `-service`）保持前台进程行为，用于开发。

### UI

```bash
# 安装依赖
cd app
npm install

# 开发模式（启动 Vite 开发服务器）
npm run dev

# TypeScript 类型检查
npm run lint

# 生产构建
npm run build

# Tauri 桌面应用
npx tauri dev
npx tauri build
```

UI 开发服务器默认监听 `http://127.0.0.1:5173`。

### 环境变量

参见 `.env.example`，支持：

- `LOCAL_SERVICE_PANEL_DATA` — 覆盖数据目录
- `LOCAL_SERVICE_PANEL_DEV_TOKEN` — 开发环境 token（方便调试）

## Pi Prompt Templates

本项目提供项目级 Pi 提示词模板，位于：

```text
.pi/prompts/
```

在 Pi 中可直接使用：

```text
/plan-next          # 阅读上下文，规划 TODO 下一项任务，但不写代码
/do-next            # 执行 TODO 下一项任务，并更新 TODO
/do-next-reviewed   # 执行 TODO 下一项任务，完成后自检并修复问题
/task <任务描述>    # 按项目流程执行指定任务
/review-task        # 按项目约定审查当前改动
```

如果 Pi 已经启动，新增或修改模板后可执行：

```text
/reload
```

## 当前阶段

**v0.6 安装与打包** — v0.1 至 v0.5 已完成。当前实现 Agent Windows Service 安装、Tauri 打包和安装/卸载脚本，使项目可分发使用。
