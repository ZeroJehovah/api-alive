# api-alive

`api-alive` 是一个部署在 VPS 上的 Codex 模型测活 Web 服务。它运行在 Ubuntu 上，调用本机安装的 `codex` 命令，并行探测一个或多个配置模型的可用性。

当前目标部署环境是 ARM CPU 的 Ubuntu 24 VPS，并假设 VPS 上已经安装且可直接执行 `codex`。

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

打开 Web 页面：

```text
http://<vps-ip>:8080
```

默认监听地址是 `0.0.0.0:8080`。如果对公网开放，建议先通过防火墙、反向代理或访问控制限制入口。

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
  "models": ["gpt-5"],
  "model_loop_counts": {"gpt-5": 1},
  "timeout_seconds": 120,
  "codex_command": "codex",
  "listen_addr": "0.0.0.0:8080",
  "max_output_chars": 4000
}
```

字段说明：

- `models`：Web 页面展示和测活时可选择的模型名。
- `model_loop_counts`：每个模型的最大尝试次数；新增模型默认是 1，可在 Models 面板修改；任一尝试成功后立即停止该模型后续尝试。
- `timeout_seconds`：单次尝试的超时时间。
- `codex_command`：VPS 上用于调用 Codex 的命令。
- `listen_addr`：Web 服务监听地址。
- `max_output_chars`：单个测活结果最多返回的输出字符数。

## 测活逻辑

每个被选中的模型都会在独立临时目录中运行。命令形态为：

```sh
codex exec --model <model> --skip-git-repo-check --ephemeral <prompt>
```

只有当 Codex 命令成功退出，并且输出中包含所选内置短提示语的预期答案时，测活才算成功。

失败包括命令执行失败、超时和输出不符合预期。命令执行失败时，错误信息优先取 Codex 输出里的最后一条 `ERROR:` 行，其次取最后一行输出，最后才使用进程错误。
