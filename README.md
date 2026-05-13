# AiForEino

基于字节跳动开源框架 **[Eino](https://github.com/cloudwego/eino)** 的示例应用：提供 **HTTP 流式对话**（SSE）、**工具调用（数据导出）**、**Redis 会话历史**、**MySQL 业务数据** 与可选的 **Langfuse** 观测。仓库中另含 **RAG 演示代码**（`demo_rag/`）与 **Redis Stack** 向量索引相关依赖，可按需启用。

---

## 功能概览

| 能力 | 说明 |
|------|------|
| 对话与流式输出 | Gin 暴露 SSE：`/agent/api/stream`，前端见 `static/home/chat.html` |
| ReAct Agent | `agent` 包内使用 Eino `react` Agent，挂载导出工具等 |
| 导出工具 | 通过 Eino Tool 触发导出图谱（`graph/export_graph`），进度以领域事件经 HTTP 层统一写 SSE |
| 会话历史 | Redis 存储；接口 `GET /agent/api/chat/history` |
| 观测 | Langfuse 在进程启动时注册为 Eino 全局回调，请求结束在 `StreamHandler` 中 `Flush`（见 `internal/observability`） |
| 进度事件 | `pkg/progress`：工具/图谱只 `TryPublish(ProgressEvent)`，SSE 组帧仅在 Agent 侧完成 |

---

## 技术栈

- **语言**：Go（见 `go.mod` 中 `go` 版本）
- **Web**：Gin
- **LLM / 编排**：CloudWeGo Eino（chat、compose、react、callbacks）
- **存储**：Redis（会话等）、MySQL（GORM）、可选 Redis Stack（向量 / RAG 演示）
- **消息**：Kafka（`cmd/reader.go` 消费侧等）
- **配置**：Viper + `config/config.yml`

---

## 仓库结构（摘要）

```
cmd/main.go              # Web 入口：配置、DB、Langfuse 初始化、路由
cmd/reader.go            # Kafka 导出消费者等（独立 main）
agent/                   # HTTP Agent：流式对话、历史读写
agent/tool/export/     # 导出 Eino Tool + 调用导出图谱
graph/export_graph/    # 导出相关 Eino 编排与节点
internal/
  database/            # Redis / MySQL 初始化
  repository/          # 数据访问
  observability/       # Langfuse 全局注册与 Flush
pkg/
  progress/            # 进度领域事件 + SSE 编码（工具侧不拼 event: 行）
  llm/                 # 模型工厂等
router/                # Gin 路由与优雅停机
static/home/           # 聊天页等静态资源
demo_rag/              # RAG 管道演示（索引 / 检索等，与主 Web 入口相对独立）
config/                # 配置文件（勿将真实密钥提交版本库）
```

---

## 快速开始

### 1. 环境要求

- Go 工具链（与 `go.mod` 中版本一致或兼容）
- 可用的 **Redis**、**MySQL**（与 `config/config.yml` 中配置一致）
- 调用大模型所需的 **API Key** 等（按你使用的 `pkg/llm` 实现填写）

### 2. 配置

- 复制并编辑 `config/config.yml`（建议本地保留真实文件，**仓库内仅保留脱敏示例**）。
- **不要**将 API Key、数据库密码、Langfuse Secret 等提交到 Git。

### 3. 运行 Web 服务

```bash
go run ./cmd
```

默认监听 **8080**（见 `router.NewApp(8080, …)`）。

### 4. 打开聊天页

浏览器访问：

- `http://127.0.0.1:8080/agent/static/chat.html`（静态路径以 `router` 中 `Static` 挂载为准）

SSE 接口示例：

- `GET http://127.0.0.1:8080/agent/api/stream?question=你好`

历史记录：

- `GET http://127.0.0.1:8080/agent/api/chat/history`

### 5. Docker（可选）

根目录提供 `docker-compose.yaml`，可拉起应用、MySQL、Redis Stack、Kafka 等。请将敏感配置通过挂载的 `config` 目录注入，勿把生产密钥写进镜像或仓库。

---

## 架构说明（与实现相关）

1. **进度与 SSE**  
   导出图谱与工具只通过 `context` 上的 `ProgressSink` 上报 `pkg/progress.ProgressEvent`；`agent.StreamHandler` 内用 channel 收集并由**同一把锁**与模型 token 流交错写入 `ResponseWriter`，避免半包与竞态。

2. **Langfuse**  
   `internal/observability.InitLangfuseEino` 在 `main` 中执行一次，向 Eino 注册全局 callback；`StreamHandler` 在请求返回前 `FlushLangfuse`，避免在业务图谱内重复 `NewLangfuseHandler`。

3. **模块路径**  
   当前 `go.mod` 中 `module main`，因此导入路径为 `main/...`。若开源或迁移，可改为带域名的模块名并全局替换 import。

---

---

## 许可与致谢

- 编排与组件能力来自 **CloudWeGo Eino** 生态。
- 使用第三方 API 与云服务时，请遵守其服务条款与数据合规要求。

若你希望 README 中增加「环境变量覆盖配置」「OpenAPI 文档」或「部署拓扑图」，可以说明偏好后再补一节。
