# api-alive

`api-alive` 是一个部署在 VPS 上的模型测活 Web 服务。Codex 分组直接复放真实 `codex-tui` 的 Responses HTTP 请求，Claude 分组直接复放真实 Claude Code 的 Messages HTTP 请求，并行探测一个或多个配置模型的可用性。

当前目标部署环境是 ARM CPU 的 Ubuntu 24 VPS。两组测活都只需要可读取的本机登录凭据；`codex` 和 `claude` 命令仅用于刷新 User-Agent 中的版本缓存。

## 构建

在 VPS 上直接构建：

```sh
go build -o bin/api-alive ./cmd/api-alive
```

从其他机器交叉构建 Linux ARM64 二进制：

```sh
GOOS=linux GOARCH=arm64 go build -o dist/api-alive-linux-arm64 ./cmd/api-alive
```

## 运行

启动服务：

```sh
./bin/api-alive --config config.json
```

启动日志会输出包含访问 token 的 Web 地址，例如：

```text
http://<vps-ip>:8080/?token=<generated-token>
```

Web 页面和所有 HTTP API 都必须携带配置中的 token。浏览器会在通过上述地址进入后为 Fetch 和 SSE 请求自动附带 token；API 客户端可以使用 Bearer token：

```sh
curl -H 'Authorization: Bearer <token>' http://<vps-ip>:8080/api/state
```

默认监听地址是 `0.0.0.0:8080`。如果对公网开放，仍建议使用 HTTPS，并通过防火墙或反向代理限制入口。

## Web 页面能力

- 查看和保存运行参数。
- 添加、删除模型名，并调整模型顺序。
- 选择单个或多个模型。
- 对单个模型或选中的多个模型执行测活。
- 展示成功、失败、耗时、尝试次数和错误输出；日志使用独立面板展示，长输出可横向滚动查看。

## 配置

首次启动时，服务会在指定路径自动创建配置文件。后续配置通过 Web 页面修改，不需要手动编辑 `config.json`。

旧版配置中的 `codex_base_url` 和 `claude_base_url` 会在加载时自动迁移为统一的 `base_url`，缺失的 `batch_min_interval_seconds` 会补为默认值 `5`，迁移结果会立即写回配置文件。

示例：

```json
{
  "models": ["gpt-5", "sonnet"],
  "model_groups": [
    {"name": "Codex", "provider": "codex", "models": ["gpt-5"]},
    {"name": "Claude", "provider": "claude", "models": ["sonnet"]}
  ],
  "model_loop_counts": {"gpt-5": 1, "sonnet": 1},
  "timeout_seconds": 120,
  "base_url": "https://cpa.example.com",
  "batch_min_interval_seconds": 5,
  "listen_addr": "0.0.0.0:8080",
  "max_output_chars": 4000,
  "token": "<generated-token>"
}
```

字段说明：

- `models`：Web 页面展示和测活时可选择的模型名。
- `model_groups`：模型分组；每个分组通过 `provider` 指定使用 `codex` 或 `claude`，Run selected 可以跨 provider 分组并行测活。
- `model_loop_counts`：每个模型的最大尝试次数；新增模型默认是 1，可在 Models 面板修改；任一尝试成功后立即停止该模型后续尝试。
- `timeout_seconds`：单次尝试的超时时间。
- `base_url`：Codex 和 Claude 共用的 API 基础地址；Codex 请求 `/v1/responses`，Claude 请求 `/v1/messages?beta=true`。默认值为 `https://cpa.aiof.top`。
- `batch_min_interval_seconds`：同一模型测试次数大于 1 时，该模型两次请求开始时间之间的最小间隔秒数。例如设为 `5`，前一次请求 2 秒后返回，下一次会再等待 3 秒；多个模型各自独立计时，测试次数为 1 的模型不等待。默认值为 `5`，设为 `0` 可关闭。
- `listen_addr`：Web 服务监听地址。
- `max_output_chars`：单个测活结果最多返回的输出字符数。
- `token`：Web 页面和所有 HTTP API 的访问凭证；首次启动或迁移旧配置时自动生成，可在 Runtime 面板更新。

## 测活逻辑

每个被选中的模型会按所属模型分组的 provider 选择测活方式。Codex 分组不启动 `codex`、`codex exec` 或 `codex app-server`，而是把一次真实 `codex-tui` 捕获到的 `/v1/responses` 请求作为模板，替换模型、当前提示语、认证信息和本次会话 ID 后直接发送：

```text
POST <base_url>/v1/responses
```

模板保留真实客户端的 developer instructions、AGENTS.md 上下文、工具定义、请求参数以及 `Originator`、`User-Agent`、turn metadata 等请求头，以减少与 `codex-tui` 实际调用之间的差异。`User-Agent` 中的 Codex 版本通过 `codex --version` 获取，并缓存到 `~/.cache/api-alive/codex-version.json`，同时记录 `updated_at`；缓存超过一天时重新获取，获取失败则沿用上一次缓存的版本。没有可用缓存且命令执行失败时回退到抓包时的 `0.163.0`。认证优先读取 `CODEX_AUTH_FILE`，否则读取 `$CODEX_HOME/auth.json` 或 `~/.codex/auth.json` 中的 `OPENAI_API_KEY` 或 `tokens.access_token`。

Claude 分组不启动 `claude` 或 `claude --print`，而是把一次真实 Claude Code `2.1.301` 捕获到的请求作为模板：

```text
POST <base_url>/v1/messages?beta=true
```

模板保留完整的 system 指令、CLAUDE.md/环境上下文、23 个工具定义、thinking/effort 等参数，以及 Claude Code 的 Stainless、Anthropic 和 session headers。每次请求会替换模型、提示语、会话 ID、设备 ID 和 Claude 版本。认证优先读取 `ANTHROPIC_AUTH_TOKEN` 或 `ANTHROPIC_API_KEY`，否则读取 `$CLAUDE_CONFIG_DIR/settings.json` 或 `~/.claude/settings.json` 的 `env`。设备 ID 从 `~/.claude.json` 的 `userID` 读取。Claude 版本通过 `claude --version` 获取并缓存到 `~/.cache/api-alive/claude-version.json`，刷新和失败回退规则与 Codex 相同。

只有请求成功，并且输出中包含所选内置短提示语的预期答案时，测活才算成功。

失败包括 HTTP/SSE 错误、超时和输出不符合预期。直接 HTTP 请求不经过 Codex 或 Claude 客户端自身的重试配置；首个非 2xx 响应或 SSE 错误会立即写入测活结果并推送到页面，不必等到超时。若上游不返回错误也不结束响应，超时信息仍会保留首条及最后一条已收到的上游错误。
