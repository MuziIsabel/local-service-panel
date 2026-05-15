# 安装与打包

## 1. 目标

本文件定义 Local Service Panel v0.6 的安装方式、打包配置和部署策略。

## 2. 安装方式

### 2.1 Agent

Agent 以 Windows Service 方式运行，通过 `install.ps1` 脚本一键安装。

安装步骤：

1. 以管理员身份打开 PowerShell。
2. 运行 `scripts\install.ps1`。
3. 脚本将：
   - 创建数据目录 `C:\ProgramData\LocalServicePanel\`
   - 复制 Agent 二进制文件
   - 注册 Windows Service `LocalServicePanelAgent`
   - 启动 Service

### 2.2 Agent 命令行

Agent 支持以下命令行标志：

| 标志 | 说明 |
|---|---|
| `-version` | 打印版本号并退出 |
| `-data <path>` | 指定数据目录（默认 `.data/`） |
| `-service install` | 注册为 Windows Service |
| `-service uninstall` | 卸载 Windows Service |
| `-service run` | 以 Windows Service 模式运行 |

默认模式（无 `-service` 标志）保持前台进程行为，用于开发。

### 2.3 UI 桌面应用

UI 通过 Tauri 打包为 Windows MSI 安装包：

```bash
cd app
npx tauri build
```

生成的安装包位于 `app/src-tauri/target/release/bundle/msi/`。

## 3. 数据目录

生产环境默认使用 `C:\ProgramData\LocalServicePanel\`，结构如下：

```text
C:\ProgramData\LocalServicePanel\
  config\
    agent.json          # Agent 配置文件
    token               # 鉴权 token
  data\
    panel.db            # SQLite 数据库
  logs\
    agent.log           # Agent 日志
    apps\               # 自定义软件日志
```

可通过 `LOCAL_SERVICE_PANEL_DATA` 环境变量或 `-data` 标志覆盖。

## 4. 目录是否适用于开发

开发环境：

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

## 5. 构建流程

### Agent 构建

```bash
cd agent
make build          # 构建 agent.exe（版本通过 ldflags 注入）
make install        # 构建并运行 install.ps1（需要管理员权限）
make uninstall      # 运行 uninstall.ps1（需要管理员权限）
```

或直接使用 Go：

```bash
go build -ldflags "-X 'github.com/user/local-service-panel/agent/internal/version.Version=0.6.0'" -o agent.exe ./cmd/agent
```

### UI 构建

```bash
cd app
npm run lint
npm run build
npx tauri build     # 生成 MSI 安装包
```

## 6. 版本号管理

版本号在以下文件中维护，需保持同步：

| 文件 | 字段 |
|---|---|
| `agent/Makefile` | `VERSION` |
| `agent/internal/version/version.go` | `Version` |
| `app/package.json` | `version` |
| `app/src-tauri/tauri.conf.json` | `version` |
| `app/src-tauri/Cargo.toml` | `version` |

## 7. 权限要求

- Agent 安装为 Windows Service 需要管理员权限。
- Windows Service 管理（启动/停止/修改启动类型）需要管理员权限。
- HKCU Run 自启不需要管理员权限。
- HKLM 写入需要管理员权限。

## 8. Token 读取

生产环境中，Tauri UI 通过 Tauri invoke 命令从 `C:\ProgramData\LocalServicePanel\config\token` 读取 token。
开发环境中，通过 `VITE_DEV_TOKEN` 环境变量读取。
