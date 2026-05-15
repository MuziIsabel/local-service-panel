# 开发环境

## 1. 目标环境

MVP 目标平台：Windows。

开发机建议：

- Windows 10/11
- Go 1.22+
- Node.js 20+
- Rust stable，Tauri 需要
- WebView2 Runtime，Windows 11 通常已内置
- SQLite，无需单独服务

## 2. 项目目录

```text
D:\agent\local-service-panel
```

规划结构：

```text
local-service-panel/
  agent/                 # Go Agent
  app/                   # Tauri + React UI
  docs/                  # 文档
  scripts/               # 脚本
  .data/                 # 开发环境本地数据，git ignore
```

## 3. Go Agent 环境

检查 Go：

```bash
go version
```

要求 Go 1.22+。

Agent 目录：

```text
agent/
  cmd/agent/main.go          # 程序入口
  internal/api/               # HTTP API
  internal/auth/              # Token 鉴权
  internal/config/            # 配置加载
  internal/db/                # SQLite 数据库
  internal/db/migrations/     # 数据库迁移 SQL
  internal/db/repository/     # 数据访问层
  internal/logging/           # 结构化日志
  internal/version/           # 版本信息
  go.mod
```

常用命令：

```bash
cd agent
go mod tidy
go test ./...
go run ./cmd/agent -data ../.data
```

启动脚本：

```bash
bash scripts/agent-dev.sh
```

验证 healthz：

```bash
curl http://127.0.0.1:17645/api/healthz
```

## Token 鉴权

首次启动时自动在 `.data/config/token` 生成随机 token（32 字节，64 位十六进制）。

- healthz 不需要 token。
- 其他所有 API 需要 `Authorization: Bearer <token>`。
- token 不会写入日志。

查看开发环境 token：

```bash
cat .data/config/token
```

开发环境可选设置环境变量：

```bash
export LOCAL_SERVICE_PANEL_DEV_TOKEN=dev-token-for-testing-only
```

这样除真实 token 外，dev token 也会被接受。

## 4. 前端/Tauri 环境

检查 Node：

```bash
node -v
npm -v
```

检查 Rust：

```bash
rustc --version
cargo --version
```

前端目录：

```text
app/
  src/
    api/client.ts          # Agent API 客户端
    pages/Dashboard.tsx     # Dashboard 页面
    types/api.ts            # TypeScript 类型
    App.tsx                 # 根组件
    main.tsx                # 入口
  src-tauri/               # Tauri 壳
    tauri.conf.json
  package.json
  vite.config.ts
```

常用命令：

```bash
cd app
npm install
npm run lint          # TypeScript 类型检查
npm run build         # 构建前端
npm run dev           # 启动开发服务器 (localhost:5173)
```

Tauri 桌面应用：

```bash
cd app
npx tauri dev         # 启动 Tauri + 前端开发
npx tauri build       # 构建桌面应用
```

## 5. 本地数据目录

开发环境使用：

```text
.data/
  config/
    agent.json
    token
  data/
    panel.db
  logs/
    agent.log
    apps/
```

正式环境建议使用：

```text
C:\ProgramData\LocalServicePanel\
  config\
  data\
  logs\
```

## 6. 默认端口

MVP 暂定：

```text
127.0.0.1:17645
```

health check：

```text
GET http://127.0.0.1:17645/api/healthz
```

## 7. 常用验证命令

Agent healthz 验证，PowerShell：

```powershell
Invoke-RestMethod http://127.0.0.1:17645/api/healthz
```

curl：

```bash
curl http://127.0.0.1:17645/api/healthz
```

## 8. 权限说明

开发早期：

- healthz、基础 API、SQLite 可普通权限运行。
- Windows Service 启停、修改启动类型可能需要管理员权限。
- HKLM 自启修改需要管理员权限。
- HKCU Run 自启通常不需要管理员权限。

原则：

- 开发默认不要要求管理员权限。
- 只有测试系统能力时再使用管理员权限。
- 不要用关键系统服务做危险测试。

## 9. 包管理器约定

待确认：npm / pnpm / yarn。

当前默认：npm。

确认后需要同步更新：

- `TODO.md`
- `README.md`
- 本文件
- 前端初始化命令

## 10. 备注

如果本文件中的命令与实际 `go.mod`、`package.json`、Tauri 配置不一致，以实际项目脚本为准，并及时更新本文档。
