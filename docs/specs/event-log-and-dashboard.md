# 功能规格：事件日志与仪表盘增强

## 1. 目标

v0.5 的目标是提升 Local Service Panel 的可观测性：

- 记录关键操作事件。
- 提供事件查询 API。
- 在 Dashboard 展示基础统计和最近事件。
- 帮助用户理解“最近发生了什么”和“哪些服务/工具异常”。

当前项目已经具备 Windows Service 管理、Custom App 管理和自启管理能力。v0.5 需要把这些操作串成可追踪的事件流。

## 2. 非目标

v0.5 不包含：

- 复杂告警系统。
- 邮件/短信/微信通知。
- 远程审计。
- 多用户审计权限。
- 指标监控系统。
- Prometheus/OpenTelemetry 集成。
- 完整健康检查调度器，可放 v0.5 后半或后续版本。

## 3. 领域概念

### Event Log

事件日志，用于记录用户操作或系统动作。

示例：

```text
WINDOWS_SERVICE_STARTED
CUSTOM_APP_CREATED
CUSTOM_APP_AUTOSTART_CHANGED
```

### Dashboard Summary

Dashboard 汇总信息，例如：

- Agent 状态。
- Windows Service 数量。
- Custom App 数量。
- 正在运行的 Custom App 数量。
- 最近事件。
- 最近错误。

## 4. 数据模型

v0.1 已创建 `event_logs` 表：

```sql
CREATE TABLE event_logs (
  id TEXT PRIMARY KEY,
  target_id TEXT,
  target_type TEXT,
  action TEXT NOT NULL,
  status TEXT NOT NULL,
  message TEXT,
  details TEXT,
  created_at TEXT NOT NULL
);
```

v0.5 应基于该表实现 repository。

### 4.1 EventLogDTO

```ts
type EventLogDTO = {
  id: string
  targetId?: string
  targetType?: string
  action: string
  status: 'success' | 'failed' | 'info'
  message?: string
  details?: string
  createdAt: string
}
```

### 4.2 Event Action 建议

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

Agent/System：

```text
AGENT_STARTED
AGENT_STOPPED
AUTH_FAILED
```

## 5. API 设计

所有事件 API 都需要 Bearer token。

### 5.1 获取事件列表

```http
GET /api/events
```

Query：

```text
limit=100
targetId=xxx
targetType=windows_service|custom_app
action=xxx
status=success|failed|info
```

响应：

```json
{
  "data": [
    {
      "id": "uuid",
      "targetId": "windows_service:Spooler",
      "targetType": "windows_service",
      "action": "WINDOWS_SERVICE_STARTED",
      "status": "success",
      "message": "Started Windows service Spooler",
      "createdAt": "2026-05-10T12:00:00Z"
    }
  ]
}
```

### 5.2 Dashboard Summary

可选新增：

```http
GET /api/dashboard/summary
```

响应：

```json
{
  "data": {
    "agent": {
      "status": "ok",
      "version": "0.1.0"
    },
    "windowsServices": {
      "total": 160,
      "running": 80,
      "stopped": 80
    },
    "customApps": {
      "total": 3,
      "running": 1,
      "stopped": 2,
      "error": 0
    },
    "recentErrors": 0
  }
}
```

MVP 可以先不做 summary API，而是在前端分别调用已有 API 和 `/api/events` 聚合。

## 6. 后端设计

推荐新增：

```text
agent/internal/db/repository/event_log.go
agent/internal/events/events.go
```

### 6.1 Repository

建议方法：

```go
type EventLogRepo struct {}

func (r *EventLogRepo) Create(e *EventLog) error
func (r *EventLogRepo) List(filter EventFilter) ([]*EventLog, error)
```

### 6.2 Event Service

建议方法：

```go
type Service struct {}

func (s *Service) Record(ctx context.Context, event Event) error
func (s *Service) List(ctx context.Context, filter Filter) ([]Event, error)
```

事件写入失败不应阻断主操作，但应写入 Agent 日志。

### 6.3 记录点

优先记录：

- Windows Service start/stop/restart/setStartType 成功和失败。
- Custom App create/update/delete/start/stop/autostart 成功和失败。

可以暂不记录：

- 查询类操作。
- healthz。
- 服务列表刷新。

## 7. 前端设计

### 7.1 Dashboard 增强

Dashboard 增加：

- Agent 状态卡片。
- Windows Service 总数/运行中/停止数。
- Custom App 总数/运行中/异常数。
- 最近事件列表。
- 最近错误列表或错误计数。

### 7.2 Events 展示

MVP 可在 Dashboard 展示最近 10 条事件。

字段：

```text
时间 | 类型 | 动作 | 状态 | 消息
```

后续可独立做 Events 页面。

## 8. 错误处理

事件 API 错误码：

```text
EVENT_LOG_CREATE_FAILED
EVENT_LOG_QUERY_FAILED
INVALID_EVENT_FILTER
```

事件写入失败：

- 不应导致原操作失败。
- 应记录到 Agent logger。

事件查询失败：

- 返回 500 和 `EVENT_LOG_QUERY_FAILED`。

## 9. 测试策略

### 9.1 单元测试

- EventLog repository Create/List。
- Event filter。
- Event DTO 映射。
- API 查询参数解析。

### 9.2 集成测试

- 执行 Custom App create 后产生事件。
- 执行 Custom App autostart 后产生事件。
- 执行 Windows Service mock 操作后产生事件。
- `GET /api/events` 返回最近事件。

### 9.3 前端验证

- Dashboard 显示统计卡片。
- Dashboard 显示最近事件。
- 操作后刷新 Dashboard 可见事件。

## 10. 安全与隐私

必须遵守：

1. 事件详情不记录 token。
2. 事件详情不记录完整敏感环境变量。
3. 事件日志不要保存过长 stdout/stderr。
4. API 需要 Bearer token。
5. 事件中的 details 应限制长度。

## 11. 验收标准

v0.5 完成时应满足：

- [ ] 实现 event_logs repository。
- [ ] 实现事件 service。
- [ ] 关键 Custom App 操作写入事件。
- [ ] 关键 Windows Service 操作写入事件。
- [ ] 实现 `GET /api/events`。
- [ ] Dashboard 展示基础统计。
- [ ] Dashboard 展示最近事件。
- [ ] 事件写入失败不阻断主操作。
- [ ] API 文档已同步。
- [ ] TODO.md 已更新。

## 12. 实现顺序建议

1. 创建规格并同步 ROADMAP/TODO。
2. 实现 event_logs repository 和 DTO。
3. 实现 `GET /api/events`。
4. 在 Custom App 操作中写入事件。
5. 在 Windows Service 操作中写入事件。
6. Dashboard 增加统计卡片。
7. Dashboard 展示最近事件。
8. 补充 API 文档和测试。
