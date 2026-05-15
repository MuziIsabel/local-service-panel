# 系统架构

## 1. 架构目标

- UI 与系统管理能力解耦。
- 后台 Agent 稳定运行，UI 可随时打开/关闭。
- 系统级敏感操作集中在 Agent 中。
- 所有托管对象使用统一模型展示。
- 先支持 Windows，后续可扩展其他平台。

## 2. 总体架构

```text
+-----------------------------+
| Tauri Desktop UI            |
| React + TypeScript          |
+-------------+---------------+
              |
              | localhost HTTP + token
              v
+-------------+---------------+
| Go Agent                    |
| Windows Service             |
+-------------+---------------+
              |
    +---------+----------+------------------+
    |                    |                  |
    v                    v                  v
Windows SCM        Startup Sources     Custom Apps
Services           Registry/Folder     Processes
                   Task Scheduler
              |
              v
+-------------+---------------+
| SQLite + Config + Logs      |
+-----------------------------+
```

## 3. 核心组件

### 3.1 Desktop UI

职责：

- 展示 Dashboard。
- 展示服务和软件列表。
- 发起操作请求。
- 展示日志和错误。
- 管理用户输入表单。

不负责：

- 直接修改注册表。
- 直接启停系统服务。
- 直接执行管理员命令。

### 3.2 Agent

职责：

- 提供本机 API。
- 校验 token。
- 读取/写入 SQLite。
- 枚举 Windows Services。
- 启停 Windows Services。
- 管理自定义软件进程。
- 管理自启配置。
- 执行健康检查。
- 写入事件日志。

### 3.3 SQLite

职责：

- 保存自定义托管目标。
- 保存健康检查配置。
- 保存事件记录。
- 保存用户设置。
- 保存部分状态快照。

### 3.4 系统适配层

用于隔离 Windows API，方便后续跨平台：

```text
internal/platform/windows/
internal/platform/linux/      # future
internal/platform/darwin/     # future
```

MVP 可先不完全抽象，但接口设计应预留扩展空间。

## 4. 关键流程

### 4.1 UI 启动流程

```text
用户打开 UI
  ↓
读取本机 token
  ↓
请求 GET /api/healthz
  ↓
如果 Agent 正常，进入 Dashboard
  ↓
如果 Agent 未运行，提示安装/启动 Agent
```

### 4.2 获取服务列表

```text
UI 请求 GET /api/targets
  ↓
Agent 查询 Windows Services
  ↓
Agent 查询自定义软件配置
  ↓
Agent 合并为 ManagedTarget 列表
  ↓
返回给 UI
```

### 4.3 启动 Windows Service

```text
UI 点击启动
  ↓
POST /api/targets/{id}/start
  ↓
Agent 校验目标类型
  ↓
调用 Windows SCM StartService
  ↓
轮询状态变化
  ↓
写入事件日志
  ↓
返回结果
```

### 4.4 启动自定义软件

```text
UI 点击启动
  ↓
POST /api/targets/{id}/start
  ↓
Agent 读取配置
  ↓
启动进程并重定向日志
  ↓
记录 PID
  ↓
返回结果
```

### 4.5 设置开机自启

```text
UI 修改开关
  ↓
POST /api/targets/{id}/autostart
  ↓
Agent 根据目标类型选择策略
  ↓
Windows Service: 修改启动类型
Custom App: 写入/删除 HKCU Run 或计划任务
  ↓
更新数据库
```

## 5. 统一托管对象

```ts
type ManagedTarget = {
  id: string
  name: string
  type: TargetType
  status: TargetStatus
  autoStart: boolean
  executablePath?: string
  workingDir?: string
  args?: string[]
  pid?: number
  startType?: string
  health?: HealthStatus
  lastError?: string
}
```

## 6. 状态分类

```text
running     运行中
stopped     已停止
starting    启动中
stopping    停止中
error       异常
unknown     未知
```

## 7. 扩展方向

- WinSW 集成，把自定义软件注册为 Windows Service。
- Task Scheduler 完整管理。
- Docker container 管理。
- Linux systemd 支持。
- 远程 Agent 管理。
- 插件式 Target Provider。
