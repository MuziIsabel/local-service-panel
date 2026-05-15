# ADR 0001：技术栈选型

## 状态

Accepted

## 背景

项目需要构建一个个人电脑本地部署的服务管理面板，核心能力包括：

- 管理 Windows Service。
- 管理开机自启项。
- 管理用户自定义软件。
- 展示运行状态和日志。
- 未来可能支持 Docker、Linux systemd、远程管理。

该项目同时需要本机系统能力和较好的桌面 UI 体验。

## 决策

采用：

```text
Go Agent + Tauri + React + TypeScript + Ant Design + SQLite
```

其中：

- Go Agent 负责系统管理能力。
- Tauri 提供桌面应用外壳。
- React + TypeScript 实现前端界面。
- Ant Design 作为 UI 组件库。
- SQLite 存储本地配置和事件。

## 理由

### Go Agent

优点：

- 适合长期运行的后台服务。
- 编译简单，分发方便。
- 系统调用和进程管理能力成熟。
- Windows 服务管理可以通过 `golang.org/x/sys/windows` 实现。
- 后续跨平台扩展成本可控。

### Tauri + React

优点：

- 比 Electron 更轻量。
- React 生态成熟，适合快速开发管理面板。
- TypeScript 有利于维护 API 类型。
- Tauri 可打包为桌面应用。

### SQLite

优点：

- 本机应用无需外部数据库。
- 轻量稳定。
- 适合存储配置、状态、事件日志。

## 替代方案

### Rust/Tauri 全家桶

优点：体积小、性能好、安全性强。

缺点：开发复杂度更高，Windows API 和业务开发速度可能慢于 Go。

### Electron + Node.js

优点：前端开发体验好，生态丰富。

缺点：体积大，长期后台服务和 Windows 系统管理能力不如 Go/Rust 稳定。

### 纯 Web + Go Agent

优点：实现简单，开发快。

缺点：桌面体验和分发体验弱于 Tauri。

## 影响

- 项目将拆分为 Agent 和 App 两个主要部分。
- Agent 作为长期运行后台服务，需要单独安装和管理。
- UI 关闭不会影响服务管理能力。
- 前后端通过本机 API 通信，需要设计 token 认证和 CORS 策略。

## 后续关注

- Agent 权限边界。
- 本机 API 安全。
- Windows Service 安装体验。
- Tauri 与 Agent 的安装包整合。
