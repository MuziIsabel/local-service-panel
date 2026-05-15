# 功能规格：Windows Service 管理

## 1. 目标

v0.2 的目标是让 Local Service Panel 可以可视化管理 Windows Services。

用户应能够：

- 查看 Windows Service 列表。
- 查看服务运行状态。
- 查看服务启动类型。
- 查看服务基础信息，例如服务名、显示名、描述、可执行路径。
- 启动服务。
- 停止服务。
- 重启服务。
- 修改服务启动类型。
- 在前端列表中搜索、过滤和识别高风险服务。

## 2. 非目标

v0.2 不包含：

- 安装 Agent 为 Windows Service。
- 管理 Linux systemd。
- 管理 Docker 容器。
- 管理 Task Scheduler。
- 管理注册表 Run 启动项。
- 管理 Startup Folder。
- 自定义软件托管。
- 远程管理。
- 批量操作系统服务。
- 绕过权限限制强行控制系统关键服务。

## 3. 术语

### Windows Service

由 Windows Service Control Manager 管理的系统服务。

### Service Name

系统内部服务名，例如：

```text
Spooler
wuauserv
```

### Display Name

用户可读名称，例如：

```text
Print Spooler
Windows Update
```

### Start Type

服务启动类型：

```text
automatic
manual
disabled
automatic_delayed
unknown
```

### Protected Service

本项目认为高风险或关键的服务。对这类服务，前端应强提示，Agent 可拒绝或要求更严格确认。

## 4. 权限要求

不同操作的权限要求不同：

| 操作 | 普通权限 | 管理员权限 |
|---|---:|---:|
| 枚举服务 | 通常可用 | 可用 |
| 查询服务状态 | 通常可用 | 可用 |
| 查询启动类型 | 通常可用 | 可用 |
| 启动服务 | 可能失败 | 通常需要 |
| 停止服务 | 可能失败 | 通常需要 |
| 修改启动类型 | 通常失败 | 通常需要 |

MVP 策略：

- Agent 先按当前运行权限执行。
- 权限不足时返回明确错误。
- 不在 v0.2 中强制实现 UAC 提权流程。
- 后续安装为 Windows Service 后，由高权限 Agent 执行敏感操作。

## 5. 安全约束

必须遵守：

1. 所有 Windows Service 管理 API 都需要 Bearer token。
2. `/api/healthz` 仍然可以无 token 访问。
3. 不允许前端直接调用 Windows API。
4. 不默认开放远程访问。
5. 不默认监听 `0.0.0.0`。
6. 操作失败不能泄露敏感环境信息。
7. 对 Protected Service 必须有保护策略。
8. 前端危险操作必须二次确认。

## 6. 关键服务保护

### 6.1 初始保护列表

初始保护列表至少包含：

```text
WinDefend
EventLog
RpcSs
PlugPlay
SamSs
LSM
Winmgmt
DcomLaunch
SecurityHealthService
mpssvc
```

说明：不同 Windows 版本中服务名称可能不同，列表需要逐步调整。

### 6.2 MVP 保护策略

MVP 推荐策略：

- 列表中仍展示 protected 服务。
- 前端显示“受保护”标签。
- 对 protected 服务禁用停止、重启、禁用启动类型等高风险操作。
- 启动 protected 服务通常允许，但仍可提示。
- Agent 层也要校验，不能只依赖前端。

### 6.3 后续增强

后续可支持：

- 配置文件自定义保护列表。
- 不同风险等级。
- 强确认码。
- 操作审计。

## 7. DTO 设计

### 7.1 WindowsServiceDTO

```ts
type WindowsServiceDTO = {
  id: string
  serviceName: string
  displayName: string
  description?: string
  status: WindowsServiceStatus
  startType: WindowsServiceStartType
  executablePath?: string
  canStop?: boolean
  canPauseAndContinue?: boolean
  protected: boolean
  lastError?: string
}
```

### 7.2 WindowsServiceStatus

```ts
type WindowsServiceStatus =
  | 'running'
  | 'stopped'
  | 'start_pending'
  | 'stop_pending'
  | 'pause_pending'
  | 'paused'
  | 'continue_pending'
  | 'unknown'
```

### 7.3 WindowsServiceStartType

```ts
type WindowsServiceStartType =
  | 'automatic'
  | 'automatic_delayed'
  | 'manual'
  | 'disabled'
  | 'unknown'
```

### 7.4 ID 规则

统一 ID：

```text
windows_service:<serviceName>
```

示例：

```text
windows_service:Spooler
```

## 8. API 设计

所有接口除 `/api/healthz` 外都需要：

```http
Authorization: Bearer <token>
```

### 8.1 获取服务列表

```http
GET /api/windows/services
```

Query：

```text
keyword=xxx
status=running|stopped|paused|unknown
startType=automatic|manual|disabled|automatic_delayed|unknown
includeProtected=true|false
```

响应：

```json
{
  "data": [
    {
      "id": "windows_service:Spooler",
      "serviceName": "Spooler",
      "displayName": "Print Spooler",
      "description": "This service spools print jobs...",
      "status": "running",
      "startType": "automatic",
      "executablePath": "C:\\Windows\\System32\\spoolsv.exe",
      "canStop": true,
      "canPauseAndContinue": false,
      "protected": false
    }
  ]
}
```

### 8.2 获取服务详情

```http
GET /api/windows/services/{serviceName}
```

### 8.3 启动服务

```http
POST /api/windows/services/{serviceName}/start
```

响应：

```json
{
  "data": {
    "serviceName": "Spooler",
    "status": "running"
  }
}
```

### 8.4 停止服务

```http
POST /api/windows/services/{serviceName}/stop
```

### 8.5 重启服务

```http
POST /api/windows/services/{serviceName}/restart
```

### 8.6 修改启动类型

```http
PATCH /api/windows/services/{serviceName}/start-type
```

请求：

```json
{
  "startType": "automatic"
}
```

可选值：

```text
automatic
manual
disabled
automatic_delayed
```

## 9. 错误响应

统一错误格式：

```json
{
  "error": {
    "code": "WINDOWS_SERVICE_START_FAILED",
    "message": "Failed to start Windows service",
    "details": "..."
  }
}
```

建议错误码：

| code | 说明 |
|---|---|
| `WINDOWS_SERVICE_NOT_FOUND` | 服务不存在 |
| `WINDOWS_SERVICE_QUERY_FAILED` | 查询失败 |
| `WINDOWS_SERVICE_START_FAILED` | 启动失败 |
| `WINDOWS_SERVICE_STOP_FAILED` | 停止失败 |
| `WINDOWS_SERVICE_RESTART_FAILED` | 重启失败 |
| `WINDOWS_SERVICE_START_TYPE_FAILED` | 修改启动类型失败 |
| `WINDOWS_SERVICE_PROTECTED` | 服务受保护，拒绝高风险操作 |
| `WINDOWS_SERVICE_PERMISSION_DENIED` | 权限不足 |
| `INVALID_START_TYPE` | 启动类型非法 |
| `UNAUTHORIZED` | 未认证或 token 错误 |

## 10. Agent 内部设计建议

推荐模块：

```text
agent/internal/domain/windows_service.go
agent/internal/windowsservice/
  provider.go
  provider_windows.go
  protected.go
```

建议接口：

```go
type Provider interface {
    List(ctx context.Context, filter Filter) ([]Service, error)
    Get(ctx context.Context, serviceName string) (*Service, error)
    Start(ctx context.Context, serviceName string) error
    Stop(ctx context.Context, serviceName string) error
    Restart(ctx context.Context, serviceName string) error
    SetStartType(ctx context.Context, serviceName string, startType StartType) error
}
```

HTTP handler 不应直接堆 Windows API 细节。

## 11. 前端设计

新增页面：

```text
app/src/pages/Services.tsx
```

列表字段：

```text
显示名 | 服务名 | 状态 | 启动类型 | 受保护 | 操作
```

基础能力：

- 搜索服务名/显示名。
- 按状态过滤。
- 按启动类型过滤。
- 状态标签。
- 受保护标签。
- 刷新按钮。

操作按钮：

- 启动
- 停止
- 重启
- 修改启动类型

危险操作：

- 停止服务
- 重启服务
- 禁用服务

这些操作前端必须二次确认。

## 12. 测试策略

### 12.1 单元测试

可测试：

- DTO 映射。
- startType/status 字符串转换。
- protected service 判断。
- API 参数校验。
- 错误码映射。

### 12.2 集成/手动测试

Windows Service 操作以手动验证为主。

验证项：

- 枚举服务列表。
- 查询服务详情。
- 搜索过滤。
- 尝试启动一个安全测试服务。
- 尝试停止一个安全测试服务。
- 尝试修改启动类型并恢复。
- protected 服务高风险操作被拒绝或禁用。

禁止使用关键系统服务做危险测试。

### 12.3 非 Windows 环境

如果在非 Windows 环境运行：

- Windows Service API 应返回明确错误，例如 `WINDOWS_SERVICE_UNSUPPORTED_PLATFORM`。
- 不应导致 Agent 启动失败。

## 13. 验收标准

v0.2 完成时应满足：

- [x] 后端可以枚举 Windows Services。
- [x] 后端可以查询服务状态和启动类型。
- [x] 后端可以查询服务详情。
- [x] 后端可以启动安全测试服务。
- [x] 后端可以停止安全测试服务。
- [x] 后端可以重启安全测试服务。
- [x] 后端可以修改启动类型并恢复。
- [x] 受保护服务高风险操作被拒绝或禁用。
- [x] 前端有服务列表页面。
- [x] 前端支持搜索、过滤、状态标签。
- [x] 前端操作按钮调用对应 API。
- [x] 危险操作有确认。
- [x] API 文档已同步。
- [x] TODO.md 已更新。

## 13.1 实现状态

状态：Done

完成记录：

- 已实现 Windows Service 领域模型、DTO、Provider 接口。
- 已实现 Windows SCM Provider，非 Windows 平台提供 unsupported stub。
- 已实现服务列表、详情、启动、停止、重启、修改启动类型 API。
- 已实现 protected service 保护策略。
- 已实现前端 Services 页面，支持搜索、过滤、标签和操作按钮。
- 前端 `npm run lint` 和 `npm run build` 已验证通过。
- Go 测试在用户本机记录为通过；当前评审环境中 Go 不在 PATH，无法复验。

## 14. 实现顺序建议

推荐拆分：

1. 定义 DTO、领域模型、Provider 接口。
2. 实现只读列表 API：`GET /api/windows/services`。
3. 实现前端服务列表页面，只展示不操作。
4. 实现服务详情 API。
5. 实现启动/停止/重启 API。
6. 实现启动类型修改 API。
7. 前端接入操作按钮和确认弹窗。
8. 补充测试和文档。
