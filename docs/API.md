# API 设计

## 1. 基础约定

MVP 使用本机 HTTP API：

```text
Base URL: http://127.0.0.1:17645
```

所有非 healthz 请求都需要：

```http
Authorization: Bearer <token>
```

请求和响应均使用 JSON。

## 2. 通用响应

成功响应直接返回数据：

```json
{
  "data": {}
}
```

错误响应：

```json
{
  "error": {
    "code": "SERVICE_START_FAILED",
    "message": "Failed to start service",
    "details": "..."
  }
}
```

### 通用错误码

所有 API 可能返回以下通用错误码：

| HTTP 状态码 | code | 说明 |
|---|---|---|
| 400 | `INVALID_JSON` | 请求体不是合法的 JSON |
| 401 | `UNAUTHORIZED` | 缺少或无效的 Authorization header |
| 404 | `NOT_FOUND` | 请求的资源不存在 |

## 3. Health Check

### GET /api/healthz

用于 UI 检查 Agent 是否运行。

响应：

```json
{
  "data": {
    "status": "ok",
    "version": "0.6.0"
  }
}
```

## 4. Targets（已规划，下一版本实现）

> 以下 `/api/targets/*` 接口已规划但尚未实现。当前请使用对应的专用 API（Windows Services / Custom Apps）。

### GET /api/targets

获取统一托管对象列表。

Query：

```text
type=windows_service|custom_app
status=running|stopped|error|unknown
keyword=xxx
```

响应：

```json
{
  "data": [
    {
      "id": "windows_service:Spooler",
      "name": "Print Spooler",
      "type": "windows_service",
      "status": "running",
      "autoStart": true,
      "executablePath": "C:\\Windows\\System32\\spoolsv.exe",
      "startType": "automatic"
    }
  ]
}
```

### GET /api/targets/{id}

获取单个托管对象详情。

### POST /api/targets/{id}/start

启动目标。

响应：

```json
{
  "data": {
    "id": "windows_service:Spooler",
    "status": "running"
  }
}
```

### POST /api/targets/{id}/stop

停止目标。

### POST /api/targets/{id}/restart

重启目标。

### POST /api/targets/{id}/autostart

修改自启状态。

请求：

```json
{
  "enabled": true
}
```

## 5. Windows Services

如果统一 targets API 不够，也可以保留服务专用 API。

### GET /api/windows/services

枚举 Windows Services。

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

### GET /api/windows/services/{serviceName}

获取 Windows Service 详情。

响应：

```json
{
  "data": {
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
}
```

### POST /api/windows/services/{serviceName}/start

启动服务。

### POST /api/windows/services/{serviceName}/stop

停止服务。

### POST /api/windows/services/{serviceName}/restart

重启服务。

### PATCH /api/windows/services/{serviceName}/start-type

修改启动类型。

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

Windows Service 状态可选值：

```text
running
stopped
start_pending
stop_pending
pause_pending
paused
continue_pending
unknown
```

Windows Service 相关错误码：

```text
WINDOWS_SERVICE_NOT_FOUND
WINDOWS_SERVICE_QUERY_FAILED
WINDOWS_SERVICE_START_FAILED
WINDOWS_SERVICE_STOP_FAILED
WINDOWS_SERVICE_RESTART_FAILED
WINDOWS_SERVICE_START_TYPE_FAILED
WINDOWS_SERVICE_PROTECTED
WINDOWS_SERVICE_PERMISSION_DENIED
WINDOWS_SERVICE_UNSUPPORTED_PLATFORM
INVALID_START_TYPE
```

## 6. Custom Apps

### GET /api/custom-apps

获取自定义软件列表。

Query：

```text
keyword=xxx
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

### GET /api/custom-apps/{id}

获取自定义软件详情。

### POST /api/custom-apps

添加自定义软件。

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

响应状态码：`201 Created`。

### PATCH /api/custom-apps/{id}

编辑自定义软件。

请求示例：

```json
{
  "name": "My App Updated",
  "workingDir": "D:\\apps\\my-app"
}
```

### DELETE /api/custom-apps/{id}

删除自定义软件。

默认仅删除管理配置，不删除真实文件。运行中的 Custom App 会拒绝删除。

### POST /api/custom-apps/{id}/start

启动自定义软件。

响应：

```json
{
  "data": {
    "id": "uuid",
    "name": "My App",
    "type": "custom_app",
    "status": "running",
    "pid": 12345,
    "executablePath": "D:\\apps\\my-app\\app.exe",
    "args": ["--port", "8080"],
    "autoStart": false,
    "createdAt": "2026-05-10T12:00:00Z",
    "updatedAt": "2026-05-10T12:05:00Z"
  }
}
```

### POST /api/custom-apps/{id}/stop

停止自定义软件。

### POST /api/custom-apps/{id}/autostart

设置自定义软件开机自启。

请求：

```json
{
  "enabled": true
}
```

响应：

```json
{
  "data": {
    "id": "uuid",
    "name": "My App",
    "type": "custom_app",
    "status": "stopped",
    "executablePath": "D:\\apps\\my-app\\app.exe",
    "args": ["--port", "8080"],
    "autoStart": true,
    "createdAt": "2026-05-10T12:00:00Z",
    "updatedAt": "2026-05-10T12:05:00Z"
  }
}
```

行为：

- `enabled=true`：写入 HKCU Run，并更新 `managed_targets.auto_start`。
- `enabled=false`：删除 HKCU Run 项，并更新 `managed_targets.auto_start`。

### GET /api/custom-apps/{id}/logs

读取自定义软件 stdout/stderr 日志。

Query：

```text
lines=200
```

响应：

```json
{
  "data": {
    "stdout": ["line1", "line2"],
    "stderr": []
  }
}
```

Custom App 状态可选值：

```text
running
stopped
error
unknown
```

Custom App 相关错误码：

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
VALIDATION_ERROR
```

Custom App 自启相关错误码：

```text
AUTOSTART_UNSUPPORTED_PLATFORM
AUTOSTART_REGISTRY_OPEN_FAILED
AUTOSTART_REGISTRY_WRITE_FAILED
AUTOSTART_REGISTRY_DELETE_FAILED
AUTOSTART_INVALID_TARGET
AUTOSTART_EXECUTABLE_NOT_FOUND
AUTOSTART_COMMAND_BUILD_FAILED
```

## 7. Logs（已规划，下一版本实现）

> 该统一日志接口已规划但尚未实现。当前自定义软件日志请使用 `GET /api/custom-apps/{id}/logs`。

### GET /api/targets/{id}/logs

Query：

```text
stream=false
lines=200
```

响应：

```json
{
  "data": {
    "stdout": ["line1", "line2"],
    "stderr": []
  }
}
```

实时日志后续可使用：

- Server-Sent Events
- WebSocket
- 轮询

## 8. Events

### GET /api/events

获取事件列表。

Query：

```text
limit=100             # 返回条数上限（默认 100）
targetId=xxx          # 按目标 ID 过滤
targetType=xxx        # 按目标类型过滤（如 custom_app、windows_service）
action=xxx            # 按动作过滤（如 CUSTOM_APP_STARTED）
status=success|failed # 按状态过滤
```

### 事件动作

Windows Service：

```text
WINDOWS_SERVICE_STARTED
WINDOWS_SERVICE_START_FAILED
WINDOWS_SERVICE_STOPPED
WINDOWS_SERVICE_STOP_FAILED
WINDOWS_SERVICE_RESTARTED
WINDOWS_SERVICE_RESTART_FAILED
WINDOWS_SERVICE_START_TYPE_CHANGED
WINDOWS_SERVICE_START_TYPE_CHANGE_FAILED
```

Custom App：

```text
CUSTOM_APP_CREATED
CUSTOM_APP_CREATE_FAILED
CUSTOM_APP_UPDATED
CUSTOM_APP_UPDATE_FAILED
CUSTOM_APP_DELETED
CUSTOM_APP_DELETE_FAILED
CUSTOM_APP_STARTED
CUSTOM_APP_START_FAILED
CUSTOM_APP_STOPPED
CUSTOM_APP_STOP_FAILED
CUSTOM_APP_AUTOSTART_CHANGED
CUSTOM_APP_AUTOSTART_CHANGE_FAILED
```

## 9. Settings（已规划，下一版本实现）

> 以下设置接口已规划但尚未实现。

### GET /api/settings

获取设置。

### PATCH /api/settings

修改设置。
