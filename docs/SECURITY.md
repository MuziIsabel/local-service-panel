# 安全与权限设计

## 1. 安全目标

- 默认只允许本机访问。
- UI 不直接执行高权限系统操作。
- 高权限能力集中在 Agent。
- 防止任意网页调用本机 API。
- 防止 token、环境变量、命令参数中的敏感信息泄漏。

## 2. 权限模型

### 2.1 Agent

Agent 作为 Windows Service 运行，可使用管理员权限或 LocalSystem 权限，负责：

- 启停 Windows Service。
- 修改服务启动类型。
- 写入注册表自启项。
- 创建计划任务，后续。
- 启动/停止自定义软件。
- 写入 ProgramData 数据目录。

### 2.2 UI

UI 以普通用户权限运行，负责：

- 展示数据。
- 收集用户操作。
- 调用 Agent API。

UI 不负责：

- 直接调用 Windows SCM。
- 直接修改 HKLM 注册表。
- 直接执行管理员命令。

## 3. API 暴露策略

MVP：

```text
host: 127.0.0.1
port: 17645
```

禁止监听：

```text
0.0.0.0
公网 IP
局域网 IP
```

除非后续明确实现远程访问认证模型。

## 4. Token 认证

安装或首次启动时生成随机 token。

要求：

- 至少 32 字节随机值。
- 存储在受限权限文件中。
- 请求头使用 Bearer token。
- token 不写入普通日志。

请求示例：

```http
Authorization: Bearer <token>
```

## 5. CORS 策略

默认不允许任意 Origin。

允许：

- Tauri 应用来源。
- 明确配置的 localhost 开发地址。

拒绝：

- 任意网页来源。
- 空泛的 `Access-Control-Allow-Origin: *`。

## 6. 高危操作确认

以下操作前端必须二次确认：

- 停止关键系统服务。
- 禁用自启。
- 删除自定义软件配置。
- 修改服务启动类型为 disabled。
- 强制结束进程树。

Agent 可维护关键服务保护列表，禁止或警告操作。

## 7. 命令执行安全

自定义软件管理涉及执行用户配置命令，必须注意：

- 启动程序尽量使用可执行路径 + 参数数组，而不是 shell 字符串。
- 默认不通过 `cmd.exe /c` 执行。
- 参数单独存储，避免命令注入。
- 工作目录必须存在。
- 可执行文件路径必须存在。

## 8. 日志安全

日志中避免输出：

- token
- password
- secret
- api key
- 完整环境变量
- 过长命令行中的敏感参数

可以对疑似敏感字段脱敏：

```text
password=******
token=******
```

## 9. 文件权限

建议目录：

```text
C:\ProgramData\LocalServicePanel\
```

文件权限：

- token 文件仅 Agent 和当前用户可读。
- 数据库文件不应被普通低权限进程随意修改。
- 日志目录可读，但需避免敏感信息。

## 10. 后续增强

- 使用 Windows Named Pipe 替代 HTTP。
- Named Pipe ACL 限制调用方。
- UI 与 Agent 进行挑战响应认证。
- 操作审计日志。
- 关键操作需要 Windows UAC 确认。
