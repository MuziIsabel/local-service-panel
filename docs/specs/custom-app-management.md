# 功能规格：自定义软件管理

## 1. 目标

v0.3 的目标是让用户可以把自己的软件、脚本或命令添加到 Local Service Panel 中进行统一管理。

用户应能够：

- 添加自定义软件配置。
- 查看自定义软件列表。
- 编辑自定义软件配置。
- 删除自定义软件配置。
- 启动自定义软件。
- 停止自定义软件。
- 查看运行状态和 PID。
- 查看 stdout/stderr 日志文件。
- 在前端通过表单完成基础管理。

## 2. 非目标

v0.3 不包含：

- 开机自启管理，放到 v0.4。
- WinSW/NSSM 注册为 Windows Service，放到后续版本。
- Docker 管理。
- 远程管理。
- 插件系统。
- 复杂进程守护和失败自动重启。
- 多用户权限系统。
- 云同步。

## 3. 术语

### Custom App

用户主动添加到本面板管理的软件、脚本或命令。

### Managed Target

统一托管对象。Custom App 是 managed_targets 表中的一种类型：

```text
custom_app
```

### Runtime State

运行时状态，例如 PID、是否运行、启动时间。运行时状态可能会过期，Agent 重启后必须校验 PID 是否仍然有效。

## 4. 数据模型

v0.1 已创建 `managed_targets` 和 `process_runtime` 表。v0.3 应基于现有表实现，不应重复建表，除非发现 schema 不足。

### 4.1 managed_targets

核心字段：

```text
id
name
type
executable_path
working_dir
args_json
start_command
stop_command
auto_start
health_check_json
pid
last_started_at
last_stopped_at
last_error
created_at
updated_at
```

v0.3 中：

- `type` 固定为 `custom_app`。
- 优先使用 `executable_path + args_json`。
- `start_command` 可暂不使用，避免 shell 注入。
- `stop_command` 可选，MVP 可先保存但不执行，或明确执行策略。
- `auto_start` 暂存配置但不实现真实自启，真实自启放 v0.4。

### 4.2 process_runtime

用于保存运行时 PID 和状态：

```text
target_id
pid
status
started_at
updated_at
```

Agent 启动后应校验 PID 是否仍存在，不能完全信任数据库中的 PID。

## 5. DTO 设计

### 5.1 CustomAppDTO

```ts
type CustomAppDTO = {
  id: string
  name: string
  type: 'custom_app'
  status: 'running' | 'stopped' | 'error' | 'unknown'
  executablePath: string
  workingDir?: string
  args: string[]
  autoStart: boolean
  pid?: number
  lastStartedAt?: string
  lastStoppedAt?: string
  lastError?: string
  createdAt: string
  updatedAt: string
}
```

### 5.2 CreateCustomAppRequest

```ts
type CreateCustomAppRequest = {
  name: string
  executablePath: string
  workingDir?: string
  args?: string[]
  autoStart?: boolean
  stopCommand?: string
}
```

### 5.3 UpdateCustomAppRequest

```ts
type UpdateCustomAppRequest = Partial<CreateCustomAppRequest>
```

## 6. API 设计

所有 Custom App API 都需要：

```http
Authorization: Bearer <token>
```

### 6.1 获取列表

```http
GET /api/custom-apps
```

Query：

```text
keyword=xxx
status=running|stopped|error|unknown
```

响应：

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "My App",
      "type": "custom_app",
      "status": "stopped",
      "executablePath": "D:\\apps\\my-app\\app.exe",
      "workingDir": "D:\\apps\\my-app",
      "args": ["--port", "8080"],
      "autoStart": false,
      "createdAt": "2026-05-10T12:00:00Z",
      "updatedAt": "2026-05-10T12:00:00Z"
    }
  ]
}
```

### 6.2 获取详情

```http
GET /api/custom-apps/{id}
```

### 6.3 添加

```http
POST /api/custom-apps
```

请求：

```json
{
  "name": "My App",
  "executablePath": "D:\\apps\\my-app\\app.exe",
  "workingDir": "D:\\apps\\my-app",
  "args": ["--port", "8080"],
  "autoStart": false
}
```

### 6.4 编辑

```http
PATCH /api/custom-apps/{id}
```

### 6.5 删除

```http
DELETE /api/custom-apps/{id}
```

默认仅删除管理配置，不删除真实文件。

如果应用正在运行，MVP 推荐拒绝删除，并提示用户先停止。

### 6.6 启动

```http
POST /api/custom-apps/{id}/start
```

响应：

```json
{
  "data": {
    "id": "uuid",
    "status": "running",
    "pid": 12345
  }
}
```

### 6.7 停止

```http
POST /api/custom-apps/{id}/stop
```

响应：

```json
{
  "data": {
    "id": "uuid",
    "status": "stopped"
  }
}
```

### 6.8 日志

```http
GET /api/custom-apps/{id}/logs?stream=false&lines=200
```

MVP 可先实现读取最近 N 行，不做实时 stream。

## 7. 进程启动策略

必须遵守：

1. 优先使用 `executablePath + args[]` 启动。
2. 不默认通过 shell 字符串执行。
3. `executablePath` 必须存在。
4. 如果设置了 `workingDir`，该目录必须存在。
5. stdout/stderr 必须重定向到日志文件。
6. 启动成功后记录 PID。
7. 启动失败时记录 `last_error` 和事件日志。

Go 实现建议：

```go
cmd := exec.CommandContext(ctx, executablePath, args...)
cmd.Dir = workingDir
```

不要默认使用：

```go
cmd.exe /c
sh -c
```

## 8. 停止策略

MVP 停止顺序：

1. 如果配置了明确可用的 stopCommand，可在后续版本支持。
2. v0.3 优先按 PID 停止。
3. 停止后校验进程是否退出。
4. 如果无法停止，返回明确错误。
5. 强制 kill 进程树属于危险操作，前端必须确认；MVP 可先不实现进程树 kill。

Windows 上可以先使用：

- `os.Process.Kill()`，简单但不处理子进程。
- 后续改进为 Windows Job Object 或 taskkill 进程树。

## 9. 日志策略

日志目录：

```text
.data/logs/apps/{target_id}/
  stdout.log
  stderr.log
```

正式环境：

```text
C:\ProgramData\LocalServicePanel\logs\apps\{target_id}\
```

要求：

- stdout 和 stderr 分开保存。
- 日志文件不存在时返回空列表，不视为错误。
- 读取日志时限制最大行数，避免一次返回过大。
- 日志中不应额外写入 token、密码、secret。

## 10. 状态校验

状态来源优先级：

1. 实际进程是否存在。
2. process_runtime 表。
3. managed_targets 中的 last_error。

Agent 启动或查询列表时，应校验 PID 是否仍然存活。

如果 PID 不存在：

- 状态应修正为 `stopped` 或 `unknown`。
- process_runtime 应更新。

## 11. 安全约束

必须遵守：

1. API 需要 Bearer token。
2. 不通过 shell 字符串执行用户输入。
3. 不删除用户真实软件文件。
4. 删除 Custom App 只删除管理配置。
5. 启动前校验 executablePath。
6. 工作目录必须存在。
7. 危险操作需要前端确认。
8. 日志和错误响应不应泄露敏感环境变量。
9. autoStart 字段 v0.3 只保存，不执行真实自启。

## 12. 前端设计

新增页面：

```text
app/src/pages/CustomApps.tsx
```

功能：

- 列表展示。
- 添加按钮。
- 编辑按钮。
- 删除按钮。
- 启动按钮。
- 停止按钮。
- 查看日志按钮。

列表字段：

```text
名称 | 状态 | PID | 可执行路径 | 工作目录 | 开机自启 | 操作
```

表单字段：

- 名称
- 可执行文件路径
- 工作目录
- 参数，一行一个或空格分隔，具体交互待定
- 是否开机自启，v0.3 仅保存配置

危险操作：

- 删除配置
- 停止进程
- 强制结束，后续

必须二次确认。

## 13. 测试策略

### 13.1 单元测试

- DTO 映射。
- args JSON 编解码。
- executablePath 校验。
- workingDir 校验。
- repository CRUD。
- API 请求校验。
- 日志 tail 函数。

### 13.2 集成测试

- 添加 Custom App。
- 编辑 Custom App。
- 删除 Custom App。
- 启动一个测试程序。
- 停止一个测试程序。
- stdout/stderr 写入日志。

### 13.3 测试程序

建议准备一个测试脚本或小程序：

- 持续运行。
- 每秒输出日志。
- 可安全停止。
- 不需要管理员权限。

## 14. 错误码

建议错误码：

```text
CUSTOM_APP_NOT_FOUND
CUSTOM_APP_CREATE_FAILED
CUSTOM_APP_UPDATE_FAILED
CUSTOM_APP_DELETE_FAILED
CUSTOM_APP_INVALID_EXECUTABLE
CUSTOM_APP_INVALID_WORKING_DIR
CUSTOM_APP_ALREADY_RUNNING
CUSTOM_APP_NOT_RUNNING
CUSTOM_APP_START_FAILED
CUSTOM_APP_STOP_FAILED
CUSTOM_APP_LOG_READ_FAILED
CUSTOM_APP_RUNNING_DELETE_DENIED
```

## 15. 验收标准

v0.3 完成时应满足：

- [x] 可以添加 Custom App。
- [x] 可以查看 Custom App 列表。
- [x] 可以编辑 Custom App。
- [x] 可以删除未运行的 Custom App 配置。
- [x] 删除操作不会删除真实文件。
- [x] 可以启动 Custom App。
- [x] 可以停止 Custom App。
- [x] 可以记录 PID。
- [x] 可以查看 stdout/stderr 日志。
- [ ] Agent 重启后能校验 PID 是否仍有效。
- [x] 前端有 Custom Apps 页面。
- [x] 前端有添加/编辑表单。
- [x] 危险操作有确认。
- [x] API 文档已同步。
- [x] TODO.md 已更新。

## 15.1 实现状态

状态：Done

完成记录：

- 已实现 Custom App 领域模型、DTO、repository 扩展和业务 Service。
- 已实现 Custom App 列表、详情、创建、编辑、删除、启动、停止、日志 API。
- 已实现进程启动、PID 记录、stdout/stderr 日志写入和读取。
- 已实现前端 Custom Apps 页面、添加/编辑表单、启动/停止/删除/日志查看操作。
- 前端 `npm run lint` 和 `npm run build` 已验证通过。
- Go 测试在用户本机记录为通过；当前评审环境中 Go 不在 PATH，无法复验。

已知技术债：

- PID 存活校验仍需加强。当前状态主要基于数据库 PID/lastError 判断，Agent 重启后可能存在 PID 过期问题。
- 停止逻辑目前使用 `os.Process.Kill()`，不是优雅停止，也不处理完整进程树。
- 日志 tail 当前读取全文件后截取最后 N 行，日志很大时需要优化。
- `autoStart` 字段已保存，但真实开机自启在 v0.4 实现。

## 16. 实现顺序建议

推荐拆分：

1. 创建功能规格并同步 ROADMAP/TODO。
2. 完善 Custom App repository 和领域 DTO。
3. 实现添加/编辑/删除/列表/详情 API，不启动真实进程。
4. 实现启动进程和 PID 记录。
5. 实现停止进程。
6. 实现 stdout/stderr 日志写入和读取。
7. 创建前端 Custom Apps 页面和表单。
8. 接入启动/停止/删除操作。
9. 补充测试和文档。
