# 功能规格：自启管理 MVP

## 1. 目标

v0.4 的目标是支持基础开机自启管理，让用户可以在面板中查看和控制：

- Windows Service 的自启状态。
- Custom App 的开机自启状态。

MVP 重点：

- Windows Service 自启通过服务启动类型体现。
- Custom App 自启通过当前用户的 HKCU Run 注册表项实现。
- 前端提供统一的自启开关或入口。

## 2. 非目标

v0.4 不包含：

- HKLM Run 管理。
- Task Scheduler 管理。
- Startup Folder 管理。
- WinSW/NSSM 注册为系统服务。
- 多用户自启管理。
- 远程机器自启管理。
- 复杂启动依赖编排。
- 失败自动重启。

## 3. 术语

### AutoStart

开机或登录后自动启动。不同目标类型实现方式不同。

### Windows Service AutoStart

Windows Service 的启动类型为以下值时，视为自启开启：

```text
automatic
automatic_delayed
```

启动类型为以下值时，视为自启关闭：

```text
manual
disabled
```

注意：`disabled` 比 `manual` 风险更高，前端需要明确提示。

### Custom App AutoStart

v0.4 MVP 使用当前用户注册表 Run 项：

```text
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
```

每个 Custom App 使用固定命名规则：

```text
LocalServicePanel_CustomApp_<id>
```

## 4. 数据模型

### 4.1 managed_targets.auto_start

Custom App 的自启偏好保存到：

```text
managed_targets.auto_start
```

v0.3 已保存该字段，v0.4 开始让它与 HKCU Run 同步。

### 4.2 Windows Service

Windows Service 自启状态不落库，以系统服务启动类型为准。

### 4.3 StartupEntry DTO

可选：如果实现统一自启页面，可定义：

```ts
type StartupEntryDTO = {
  id: string
  name: string
  targetType: 'windows_service' | 'custom_app'
  targetId: string
  enabled: boolean
  source: 'windows_service' | 'hkcu_run'
  command?: string
  protected?: boolean
}
```

## 5. API 设计

所有自启相关 API 都需要：

```http
Authorization: Bearer <token>
```

## 5.1 Windows Service 自启

可以复用已有接口：

```http
PATCH /api/windows/services/{serviceName}/start-type
```

映射关系：

| enabled | startType |
|---|---|
| true | automatic |
| false | manual |

如果用户选择延迟自启，可使用：

```text
automatic_delayed
```

不建议把“关闭自启”默认映射为 `disabled`，因为风险较高。

## 5.2 Custom App 设置自启

推荐新增接口：

```http
POST /api/custom-apps/{id}/autostart
```

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
    "autoStart": true
  }
}
```

行为：

- `enabled=true`：写入 HKCU Run，并更新 `managed_targets.auto_start = 1`。
- `enabled=false`：删除 HKCU Run 项，并更新 `managed_targets.auto_start = 0`。

## 5.3 自启列表，可选

如果实现统一自启页：

```http
GET /api/startup
```

响应：

```json
{
  "data": [
    {
      "id": "custom_app:uuid",
      "name": "My App",
      "targetType": "custom_app",
      "targetId": "uuid",
      "enabled": true,
      "source": "hkcu_run",
      "command": "\"D:\\tools\\my-tool.exe\" --port 8080"
    }
  ]
}
```

MVP 可以先不做统一 `/api/startup`，只在 Custom Apps 页面和 Services 页面提供自启控制。

## 6. Custom App Run 命令构造

必须正确处理路径和参数。

示例：

```text
"D:\tools\my-tool.exe" --port 8080
```

规则：

1. executablePath 必须用双引号包裹。
2. 参数中如果包含空格，也必须正确引用。
3. 不通过 shell 字符串执行，只是注册启动命令。
4. workingDir 无法直接通过 HKCU Run 表达。MVP 可接受该限制，并在文档中说明。

注意：如果 Custom App 强依赖 workingDir，HKCU Run 启动后行为可能与面板启动不同。后续可通过生成 wrapper 脚本或 Agent 自启托管解决。

## 7. Windows 注册表实现

Go 可使用：

```text
golang.org/x/sys/windows/registry
```

路径：

```text
registry.CURRENT_USER
Software\Microsoft\Windows\CurrentVersion\Run
```

操作：

- SetStringValue
- DeleteValue
- GetStringValue

非 Windows 平台：

- 返回 `AUTOSTART_UNSUPPORTED_PLATFORM`。
- Agent 不应启动失败。

## 8. 安全约束

必须遵守：

1. API 需要 Bearer token。
2. 默认只操作 HKCU，不操作 HKLM。
3. 不默认需要管理员权限。
4. 不默认开放远程访问。
5. 不为不存在的 executablePath 创建自启项。
6. 不删除非本项目命名规则的注册表项。
7. 禁用 Windows Service 不应作为默认关闭自启方式。
8. 对 protected Windows Service 禁止设置为 disabled。
9. 日志不输出 token 或敏感环境变量。

## 9. 前端设计

### 9.1 Services 页面

Windows Service 可在启动类型处提供操作：

- 设置自动
- 设置延迟自动
- 设置手动
- 设置禁用，高风险，需要强确认

MVP 已有修改启动类型弹窗，可继续使用。

### 9.2 Custom Apps 页面

Custom Apps 页面已存在 autoStart 字段和表单开关。

v0.4 需要：

- 添加列表中的 AutoStart 开关。
- 添加/编辑时同步真实 HKCU Run 状态，或保存后调用 autostart API。
- 自启开关变更需要调用：

```http
POST /api/custom-apps/{id}/autostart
```

### 9.3 统一 Startup 页面，可选

可新增：

```text
app/src/pages/Startup.tsx
```

MVP 可暂不做，除非需要集中展示。

## 10. 错误码

建议错误码：

```text
AUTOSTART_UNSUPPORTED_PLATFORM
AUTOSTART_REGISTRY_OPEN_FAILED
AUTOSTART_REGISTRY_WRITE_FAILED
AUTOSTART_REGISTRY_DELETE_FAILED
AUTOSTART_INVALID_TARGET
AUTOSTART_EXECUTABLE_NOT_FOUND
AUTOSTART_COMMAND_BUILD_FAILED
```

## 11. 测试策略

### 11.1 单元测试

- Run value name 构造。
- Run command 构造与参数引用。
- enabled 状态同步。
- unsupported platform stub。
- Custom App autostart API 参数校验。

### 11.2 手动测试

Windows 上验证：

1. 创建 Custom App。
2. 启用自启。
3. 检查 HKCU Run 项存在。
4. 禁用自启。
5. 检查 HKCU Run 项删除。
6. 修改 executablePath 或 args 后重新启用自启，命令更新。

### 11.3 重启测试

真实重启验证可放到阶段验收：

- 启用自启后重启电脑。
- 登录后确认 Custom App 启动。

## 12. 验收标准

v0.4 完成时应满足：

- [x] Custom App 可以启用自启。
- [x] Custom App 可以禁用自启。
- [x] HKCU Run 项命名符合规则。
- [x] HKCU Run 项命令正确引用路径和参数。
- [x] managed_targets.auto_start 与 HKCU Run 同步。
- [x] 非 Windows 平台返回明确 unsupported 错误。
- [x] Windows Service 自启可通过启动类型控制。
- [x] 前端 Custom Apps 页面可以切换 autoStart。
- [x] 危险操作有确认。
- [x] API 文档已同步。
- [x] TODO.md 已更新。

## 12.1 实现状态

状态：Done

完成记录：

- 已实现 `internal/autostart` 包，包含命令构造、注册表值命名、Windows HKCU Run Provider 和非 Windows stub。
- 已实现 Custom App 自启 Service 方法。
- 已实现 `POST /api/custom-apps/{id}/autostart`。
- 已实现前端 Custom Apps 页面 AutoStart 开关。
- 已通过手动 E2E 验证：启用自启写入 HKCU Run，禁用自启删除 HKCU Run。

已知限制/技术债：

- HKCU Run 无法表达 workingDir，依赖工作目录的 Custom App 自启后行为可能与面板启动不同。
- 当前仅支持 HKCU Run，不支持 HKLM、Task Scheduler、Startup Folder。
- 命令参数引用已覆盖常见空格场景，但复杂引号和特殊字符仍需后续强化。
- 尚无统一 Startup 页面，当前主要通过 Custom Apps 和 Services 页面分别管理。

## 13. 实现顺序建议

1. 创建规格并同步 ROADMAP/TODO。
2. 实现 autostart 包：命令构造、HKCU Run provider、非 Windows stub。
3. 实现 Custom App autostart Service 方法。
4. 实现 `POST /api/custom-apps/{id}/autostart`。
5. 前端 Custom Apps 页面接入 AutoStart 开关。
6. 补充 API 文档、测试和手动验证说明。
