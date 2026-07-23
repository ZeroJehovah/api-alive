# api-alive

`api-alive` 是一个部署在 VPS 上的 CLI 模型测活 Web 服务。它运行在 Ubuntu 上，调用本机安装的 `codex` 或 `claude` 命令，并行探测一个或多个配置模型的可用性。

当前目标部署环境是 ARM CPU 的 Ubuntu 24 VPS，并假设 VPS 上已经安装且可直接执行所选 provider 对应的 CLI。

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
  "codex_command": "codex",
  "claude_command": "claude",
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
- `codex_command`：VPS 上用于调用 Codex 的命令。
- `claude_command`：VPS 上用于调用 Claude Code 的命令。
- `listen_addr`：Web 服务监听地址。
- `max_output_chars`：单个测活结果最多返回的输出字符数。
- `token`：Web 页面和所有 HTTP API 的访问凭证；首次启动或迁移旧配置时自动生成，可在 Runtime 面板更新。

## 测活逻辑

每个被选中的模型都会在独立临时目录中运行，并按所属模型分组的 provider 选择 CLI。Codex 分组的命令形态为：

```sh
codex exec --model <model> --skip-git-repo-check --ephemeral <prompt>
```

Claude 分组的命令形态为：

```sh
claude --model <model> --print --no-session-persistence <prompt>
```

只有当所选 CLI 命令成功退出，并且输出中包含所选内置短提示语的预期答案时，测活才算成功。

失败包括命令执行失败、超时和输出不符合预期。命令执行失败时，错误信息优先取 CLI 输出里的最后一条 `ERROR:` 行，其次取最后一行输出，最后才使用进程错误。
