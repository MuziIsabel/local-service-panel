# 数据模型

## 1. 领域模型

### 1.1 ManagedTarget

统一托管对象。

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
  startCommand?: string
  stopCommand?: string
  healthCheck?: HealthCheckConfig
  lastStartedAt?: string
  lastStoppedAt?: string
  lastError?: string
  createdAt: string
  updatedAt: string
}
```

### 1.2 TargetType

```text
windows_service
custom_app
startup_item
scheduled_task
process
docker_container
```

MVP 实现：

```text
windows_service
custom_app
```

### 1.3 TargetStatus

```text
running
stopped
starting
stopping
error
unknown
```

### 1.4 HealthCheckConfig

```ts
type HealthCheckConfig = {
  enabled: boolean
  type: 'none' | 'http' | 'tcp' | 'process'
  url?: string
  host?: string
  port?: number
  intervalSeconds?: number
  timeoutSeconds?: number
}
```

MVP 可先只保存配置，不完整实现健康检查。

## 2. SQLite 表设计草案

### 2.1 managed_targets

用于保存用户自定义添加的软件。Windows Service 通常实时从系统读取，不全部落库；只有用户收藏、别名、备注等扩展信息才落库。

```sql
CREATE TABLE managed_targets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  executable_path TEXT,
  working_dir TEXT,
  args_json TEXT,
  start_command TEXT,
  stop_command TEXT,
  auto_start INTEGER NOT NULL DEFAULT 0,
  health_check_json TEXT,
  pid INTEGER,
  last_started_at TEXT,
  last_stopped_at TEXT,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

### 2.2 target_overrides

保存系统服务的用户自定义信息，例如收藏、备注、隐藏。

```sql
CREATE TABLE target_overrides (
  id TEXT PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_key TEXT NOT NULL,
  display_name TEXT,
  favorite INTEGER NOT NULL DEFAULT 0,
  hidden INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(target_type, target_key)
);
```

### 2.3 event_logs

记录用户操作和系统事件。

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

### 2.4 settings

保存系统设置。

```sql
CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

### 2.5 process_runtime

保存自定义软件运行时状态。该表数据不一定完全可信，Agent 启动后应校验 PID 是否仍存在。

```sql
CREATE TABLE process_runtime (
  target_id TEXT PRIMARY KEY,
  pid INTEGER,
  status TEXT NOT NULL,
  started_at TEXT,
  updated_at TEXT NOT NULL
);
```

## 3. ID 设计

- 自定义软件：UUID。
- Windows Service：`windows_service:<service_name>`。
- 注册表启动项：`registry_run:<scope>:<name>`。
- Startup 文件夹项：`startup_folder:<scope>:<filename>`。

## 4. 时间格式

统一使用 ISO 8601 UTC 字符串：

```text
2026-05-10T12:00:00Z
```

## 5. 配置文件

Agent 配置示例（JSON 格式）：

```json
{
  "server": {
    "host": "127.0.0.1",
    "port": 17645
  },
  "log": {
    "level": "info",
    "format": "json"
  }
}
```

数据库文件位置：`<dataDir>/data/panel.db`
