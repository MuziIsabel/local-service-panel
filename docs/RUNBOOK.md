# 运行与排障手册

## 1. 目标

本文件记录 Local Service Panel 常见运行问题和排障方法。随着功能实现逐步补充。

## 2. Agent 启动失败

可能原因：

- 端口被占用。
- 配置文件路径错误。
- 数据目录无权限。
- SQLite 数据库打开失败。
- Windows Service 运行权限不足。

检查项：

```bash
# 检查端口，Windows PowerShell
netstat -ano | findstr 17645
```

处理建议：

- 更换端口或停止占用进程。
- 检查 `.data/` 或 `C:\ProgramData\LocalServicePanel\` 权限。
- 查看 Agent 日志。

## 3. UI 连接不上 Agent

可能原因：

- Agent 未启动。
- 端口不一致。
- token 不一致。
- CORS 或请求头错误。
- 防火墙或安全软件拦截。

检查项：

```bash
curl http://127.0.0.1:17645/api/healthz
```

如果 healthz 不通，先排查 Agent。

如果 healthz 通但其他 API 不通，检查 token。

## 4. token 认证失败

可能原因：

- UI 读取了错误 token 文件。
- Agent 重新生成了 token。
- 开发环境配置和正式环境配置混用。
- 请求头格式错误。

正确格式：

```http
Authorization: Bearer <token>
```

注意：不要把 token 输出到日志或提交到代码库。

## 5. SQLite 打开失败

可能原因：

- 数据目录不存在。
- 文件被占用。
- 权限不足。
- 数据库损坏。

处理建议：

- 开发环境先删除 `.data/data/panel.db` 重新初始化。
- 正式环境不要直接删除数据库，先备份。
- 检查目录权限。

## 6. Windows Service 操作失败

可能原因：

- Agent 权限不足。
- 服务不存在。
- 服务状态不允许当前操作。
- 服务启动超时。
- 服务本身依赖失败。
- 被安全策略或系统保护阻止。

处理建议：

- 以管理员权限运行 Agent。
- 查看 Windows Event Viewer。
- 避免操作关键系统服务。
- 操作前记录服务原始状态。

## 7. Custom App 启动失败

可能原因：

- 可执行文件路径不存在。
- 工作目录不存在。
- 参数错误。
- 缺少环境变量。
- 端口被占用。
- 权限不足。

处理建议：

- 检查 executable path。
- 检查 working directory。
- 查看 stdout/stderr 日志。
- 在命令行手动运行同一命令确认。

## 8. Custom App 停止失败

可能原因：

- PID 已失效。
- 进程已自行退出。
- 子进程仍在运行。
- 权限不足。
- stopCommand 配置错误。

处理建议：

- 重新刷新状态。
- 检查进程树。
- 优先使用 stopCommand。
- 强制结束进程树应要求用户确认。

## 9. 自启不生效

可能原因：

- 注册表项未写入。
- 路径中空格未正确处理。
- 参数拼接错误。
- 软件需要管理员权限。
- 被安全软件拦截。

检查项：

```text
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
```

MVP 优先使用 HKCU Run，不默认写 HKLM。

## 10. 日志位置

开发环境：

```text
.data/logs/
```

正式环境：

```text
C:\ProgramData\LocalServicePanel\logs\
```

日志中不应包含：

- token
- password
- secret
- api key

## 11. 常见恢复策略

开发阶段可尝试：

1. 停止 Agent。
2. 备份 `.data/`。
3. 删除临时数据库。
4. 重新启动 Agent。
5. 重新验证 healthz。

正式环境不要直接删除数据，先备份。
