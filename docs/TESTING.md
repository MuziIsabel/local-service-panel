# 测试策略

## 1. 目标

Local Service Panel 涉及本机进程、Windows Service、注册表、自启和权限操作。测试策略需要兼顾安全、可重复和个人电脑环境差异。

## 2. 测试分层

```text
单元测试 → 集成测试 → 手动验证 → 高风险功能专项验证
```

## 3. 单元测试

适合自动化测试：

- 配置加载
- token 生成和校验
- API 响应格式
- 错误码映射
- 参数解析
- 数据模型转换
- repository 基础逻辑
- health check 纯逻辑

Go 命令：

```bash
go test ./...
```

前端命令以后以实际项目为准：

```bash
npm run lint
npm run build
```

## 4. 集成测试

适合在开发机本地验证：

- Agent HTTP server 启动
- `/api/healthz` 返回正常
- SQLite 初始化和 migration
- Custom App 配置增删改查
- 日志文件写入
- token 鉴权中间件

原则：

- 尽量使用 `.data/` 临时目录。
- 不依赖真实系统关键服务。
- 不修改 HKLM。
- 不默认需要管理员权限。

## 5. Windows Service 手动测试

Windows Service 管理涉及系统能力，MVP 阶段以手动验证为主。

测试原则：

- 不要拿关键系统服务做停止/禁用测试。
- 优先使用测试服务或安全的第三方测试服务。
- 修改启动类型后要恢复原状。
- 操作前记录原始状态。

建议测试项：

- 枚举服务列表。
- 查询服务状态。
- 查询启动类型。
- 启动一个安全测试服务。
- 停止一个安全测试服务。
- 重启一个安全测试服务。
- 修改启动类型并恢复。

禁止默认测试：

- Windows Defender 相关服务。
- EventLog。
- RpcSs。
- PlugPlay。
- SamSs。
- 网络、登录、安全核心服务。

## 6. Custom App 手动测试

可以准备一个简单测试程序或脚本，例如：

- 启动后持续运行的 Go/Node/Python 小服务。
- 监听本地端口的 HTTP echo 服务。
- 每秒输出日志的脚本。

测试项：

- 添加 Custom App。
- 启动 Custom App。
- 停止 Custom App。
- 查看 PID。
- 查看 stdout/stderr 日志。
- Agent 重启后校验 PID 是否仍有效。

## 7. 自启测试

MVP 自定义软件自启优先 HKCU Run。

测试项：

- 启用自启后注册表项存在。
- 禁用自启后注册表项删除。
- 自启命令正确处理路径和参数。
- 不写入 HKLM，除非明确需要管理员权限。

注意：真实重启验证可以放到阶段验收，不要求每轮开发都执行。

## 8. 安全测试

基础安全验证：

- 未带 token 请求受保护 API 应失败。
- 错误 token 应失败。
- `/api/healthz` 按设计可无 token。
- Agent 只监听 `127.0.0.1`。
- 日志不包含 token。
- CORS 不允许任意来源。

## 9. UI 验证

UI 侧验证：

- Agent 正常时显示连接成功。
- Agent 不可用时显示明确错误。
- API 错误时有用户可理解提示。
- 危险操作需要确认。
- 服务列表支持基础刷新。

## 10. 回归测试清单

每个较大阶段完成后至少检查：

- [ ] Agent 可以启动。
- [ ] UI 可以启动。
- [ ] UI 可以连接 Agent。
- [ ] SQLite 可以初始化。
- [ ] healthz 正常。
- [ ] TODO.md 已更新。
- [ ] 相关 docs 已同步。

## 11. 测试数据目录

开发和测试默认使用：

```text
.data/
  config/
  data/
  logs/
```

避免测试过程污染真实：

```text
C:\ProgramData\LocalServicePanel\
```

除非正在做安装包或正式运行验证。
