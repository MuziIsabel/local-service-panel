# 功能规格：Agent Healthz

## 1. 目标

实现 Agent 最小健康检查接口，用于 UI 和开发者判断 Agent 是否正在运行。

接口：

```text
GET /api/healthz
```

## 2. 非目标

本规格不包含：

- token 鉴权
- SQLite 初始化检查
- Windows Service 管理
- UI 页面实现
- 安装为 Windows Service

## 3. API

### GET /api/healthz

请求：无 body。

鉴权：MVP 中该接口允许无 token 访问。

成功响应：

```json
{
  "data": {
    "status": "ok",
    "version": "0.1.0"
  }
}
```

建议字段：

| 字段 | 说明 |
|---|---|
| status | 固定为 `ok` 表示 Agent 进程可响应 |
| version | Agent 版本 |

## 4. 监听地址

默认：

```text
127.0.0.1:17645
```

不得默认监听：

```text
0.0.0.0
```

## 5. 错误处理

healthz 本身尽量少依赖外部组件，避免因为数据库或 Windows API 暂时不可用导致 UI 无法判断 Agent 是否存活。

如果后续需要完整健康检查，可新增：

```text
GET /api/readyz
```

用于检查数据库、配置、权限等依赖状态。

## 6. 验收标准

- [ ] Agent 可以启动 HTTP server。
- [ ] 访问 `GET /api/healthz` 返回 HTTP 200。
- [ ] 响应 JSON 符合 `docs/API.md`。
- [ ] 默认监听 `127.0.0.1:17645`。
- [ ] 不需要 token。
- [ ] README 或开发文档中有验证命令。
- [ ] TODO.md 已更新。

## 7. 验证命令

curl：

```bash
curl http://127.0.0.1:17645/api/healthz
```

PowerShell：

```powershell
Invoke-RestMethod http://127.0.0.1:17645/api/healthz
```

预期返回：

```json
{
  "data": {
    "status": "ok",
    "version": "0.1.0"
  }
}
```

## 8. 后续扩展

后续可以增加：

- uptime
- build commit
- build time
- config path
- readyz
- metrics

但 MVP 不需要。 
