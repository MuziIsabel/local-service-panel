# 项目上下文

## 1. 项目一句话

Local Service Panel 是一个运行在个人电脑本机的服务管理面板，用来统一查看和管理 Windows 服务、开机自启项、自定义软件、后台进程、日志与健康状态。

## 2. 为什么做

个人电脑上的后台能力通常分散在多个地方：

- Windows Services
- 注册表 Run 启动项
- Startup 文件夹
- Task Scheduler 计划任务
- 普通后台进程
- 用户手动启动的软件
- Docker 容器，后续

这些来源状态分散、权限模型不同、开机自启方式不同。项目目标是提供统一可视化入口，让个人用户可以知道“现在有哪些东西在跑、哪些东西会开机启动、哪些东西异常、哪些东西可以安全管理”。

## 3. 核心领域语言

### 3.1 ManagedTarget

统一托管对象。任何可以在面板里展示、启动、停止、重启或配置自启的对象，都可以抽象为 ManagedTarget。

典型类型：

- `windows_service`
- `custom_app`
- `startup_item`
- `scheduled_task`
- `process`
- `docker_container`

MVP 只实现：

- `windows_service`
- `custom_app`

### 3.2 Windows Service

Windows 系统服务。它由 Windows Service Control Manager 管理。MVP 中 Windows Service 信息主要实时从系统读取，不完整落库。

允许操作：

- 查看状态
- 启动
- 停止
- 重启
- 修改启动类型

注意：部分关键系统服务需要安全保护。

### 3.3 Custom App

用户主动添加到本面板管理的软件、脚本或命令。

Custom App 需要落库，因为它是本项目自身管理的配置。

典型字段：

- 名称
- 可执行文件路径
- 工作目录
- 参数数组
- 停止命令，可选
- 是否开机自启
- 健康检查配置，可选

### 3.4 AutoStart

开机自启能力。不同 ManagedTarget 的自启实现不同：

- Windows Service：修改服务启动类型
- Custom App：MVP 优先使用 HKCU Run
- Scheduled Task：后续支持
- Startup Folder：后续支持

### 3.5 Agent

本机后台服务，负责所有系统能力和敏感操作。

Agent 是系统边界，UI 不能绕过 Agent 直接操作系统。

### 3.6 UI

Tauri + React 桌面应用。UI 只负责展示、输入和调用 Agent API。

## 4. 关键不变量

以下约束不应被随意打破：

1. UI 不直接执行系统敏感操作。
2. Agent API 默认只监听 `127.0.0.1`。
3. 非 healthz API 必须有认证机制。
4. token、密码、secret 不得写入日志。
5. Windows Service 与 Custom App 使用统一展示模型，但底层管理方式不同。
6. Windows Service 不应全部复制进 managed_targets 表；系统实时状态以系统为准。
7. Custom App 是用户配置，必须持久化。
8. TODO.md 是当前开发状态源，实际推进后必须更新。
9. API、数据模型、安全策略变化时必须同步文档。
10. MVP 优先，不为后续远程管理或插件系统提前引入复杂度。

## 5. 当前阶段边界

当前阶段：v0.1 文档与项目骨架。

应该做：

- 初始化 Go Agent
- 初始化 Tauri + React UI
- 建立健康检查接口
- 建立基础开发命令
- 保证 UI 能连接 Agent

不应该做：

- 完整 Windows Service 管理
- 完整自启扫描
- Docker 管理
- 远程管理
- 多用户权限
- 插件系统

## 6. MVP 边界

MVP 只支持 Windows。

MVP 必须支持：

- Windows Service 列表
- Windows Service 启动、停止、重启
- Windows Service 启动类型查看和修改
- 添加自定义软件
- 启动、停止自定义软件
- 自定义软件基础自启配置
- SQLite 持久化
- Tauri UI 展示

MVP 不支持：

- Linux/macOS
- Docker
- 云同步
- 远程公网访问
- 多用户系统
- 插件市场

## 7. 架构方向

推荐 Agent 内部采用 Provider/Adapter 思路。

示意：

```text
API Handler
  ↓
Application Service / Usecase
  ↓
Target Provider
  ↓
Windows SCM / Process Manager / Startup Registry / SQLite
```

未来可以有：

- WindowsServiceProvider
- CustomAppProvider
- StartupProvider
- DockerProvider
- SystemdProvider

但 MVP 不需要过度抽象，只需避免把 HTTP handler、数据库和 Windows API 强耦合在一起。

## 8. 状态刷新原则

服务状态是动态的。

MVP：

- UI 可以轮询 `GET /api/targets`
- Agent 每次请求实时获取关键状态
- Custom App 的 PID 需要验证是否仍存活

后续：

- SSE
- WebSocket
- 状态缓存
- 后台健康检查调度器

## 9. 事件日志原则

关键操作都应该逐步纳入事件日志：

- 启动服务
- 停止服务
- 重启服务
- 修改自启
- 添加 Custom App
- 启动 Custom App
- 停止 Custom App
- 操作失败
- 鉴权失败

事件日志用于 Dashboard、审计和排障。

## 10. 术语避免混淆

- Service：通常指 Windows Service。
- Target：面板中统一管理的对象。
- Custom App：用户添加的软件或脚本。
- Startup Item：系统开机自启来源中的一项。
- Agent：后台能力进程。
- UI/App：桌面界面，不等于被管理的 Custom App。
