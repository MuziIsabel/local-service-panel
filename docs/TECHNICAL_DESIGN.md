# 技术方案

## 1. 技术栈

采用以下技术栈：

```text
Agent: Go
Frontend: React + TypeScript
Desktop Shell: Tauri
UI: Ant Design
Database: SQLite
API: localhost HTTP，后续可升级为 Named Pipe
Target OS: Windows first
```

## 2. 选型理由

### 2.1 Go Agent

Go 适合编写本机后台服务：

- 编译为单文件二进制。
- 并发和后台任务实现简单。
- HTTP API、进程管理、文件管理能力成熟。
- 可通过 `golang.org/x/sys/windows` 调用 Windows 系统能力。
- 后续扩展 Linux systemd 比较自然。

### 2.2 Tauri + React

Tauri 负责桌面壳，React 负责界面：

- 打包体积比 Electron 小。
- 前端开发效率高。
- UI 体验现代化。
- 与本机 Agent 可通过 localhost API 通信。

### 2.3 SQLite

SQLite 适合作为本机配置数据库：

- 不需要额外数据库服务。
- 备份和迁移简单。
- 性能足够。
- 可以存储托管目标、状态快照、事件日志、配置等。

## 3. 运行时结构

```text
local-service-panel-agent.exe
  - 作为 Windows Service 运行
  - 监听 127.0.0.1:可配置端口
  - 提供 REST API
  - 管理系统服务、自启项、自定义软件

local-service-panel-ui.exe
  - Tauri 桌面应用
  - 普通权限运行
  - 调用 Agent API
```

## 4. Agent 模块划分

```text
agent/
  cmd/agent/              # 程序入口
  internal/api/           # HTTP API
  internal/auth/          # 本机 token 校验
  internal/config/        # 配置加载
  internal/db/            # SQLite 访问
  internal/domain/        # 领域模型
  internal/service/       # Windows Service 管理
  internal/startup/       # 开机自启管理
  internal/process/       # 进程管理
  internal/customapp/     # 自定义软件管理
  internal/health/        # 健康检查
  internal/logging/       # 日志管理
  internal/events/        # 事件记录
```

## 5. 前端模块划分

```text
app/
  src/
    api/                  # Agent API client
    components/           # 通用组件
    pages/
      Dashboard.tsx
      Services.tsx
      Startup.tsx
      CustomApps.tsx
      Logs.tsx
      Settings.tsx
    routes/               # 路由
    store/                # 状态管理
    types/                # TypeScript 类型
```

## 6. API 通信

MVP 阶段使用 HTTP：

- Agent 监听 `127.0.0.1`。
- 首次安装生成随机 token。
- UI 从本机安全配置文件读取 token。
- 请求头携带：

```http
Authorization: Bearer <token>
```

后续可升级：

- Windows Named Pipe
- Tauri command bridge
- mTLS 本地证书

## 7. Windows 能力实现

### 7.1 Windows Service

Go 包：

```text
golang.org/x/sys/windows/svc
golang.org/x/sys/windows/svc/mgr
```

支持：

- 枚举服务
- 查询状态
- 启动服务
- 停止服务
- 修改启动类型

### 7.2 开机自启

MVP 先支持：

- 自定义软件通过注册表 HKCU Run 自启。
- Windows Service 启动类型修改。

后续支持：

- HKLM Run
- Startup Folder
- Task Scheduler

### 7.3 自定义软件

MVP 以普通进程托管为主：

- Agent 启动进程。
- 记录 PID。
- stdout/stderr 重定向到日志文件。
- 停止时优先使用 stopCommand；否则结束进程树。

后续可加入 WinSW，把软件注册为系统服务。

## 8. 数据目录

建议：

```text
C:\ProgramData\LocalServicePanel\
  config\
    agent.yaml
    token
  data\
    panel.db
  logs\
    agent.log
    apps\
```

开发环境可使用：

```text
./.data/
  config/
  data/
  logs/
```

## 9. 状态模型

托管目标统一抽象为 ManagedTarget：

```text
windows_service
custom_app
startup_item
scheduled_task
process
docker_container
```

MVP 实现：

- windows_service
- custom_app

## 10. 错误处理

API 错误统一返回：

```json
{
  "error": {
    "code": "SERVICE_START_FAILED",
    "message": "Failed to start service",
    "details": "..."
  }
}
```

前端需要展示用户可理解的错误信息，并保留 details 用于诊断。
