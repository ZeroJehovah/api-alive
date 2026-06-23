# test-api-alive

`test-api-alive` is a VPS-hosted Web dashboard for probing Codex model availability. It runs on Ubuntu, calls the local `codex` command, and probes one or more configured model names in parallel.

The current target deployment is an ARM Ubuntu 24 VPS with Codex already installed.

## Build

Build on the VPS:

```sh
go build -o bin/api-alive ./cmd/api-alive
```

Cross-build a Linux ARM64 binary from another machine:

```sh
GOOS=linux GOARCH=arm64 go build -o dist/api-alive-linux-arm64 ./cmd/api-alive
```

## Run

Create or edit `config.json`, then start the server:

```sh
./bin/api-alive --config config.json
```

Open the dashboard:

```text
http://<vps-ip>:8080
```

By default the server listens on `0.0.0.0:8080`. Put it behind a firewall, reverse proxy, or access control before exposing it to the public Internet.

## Dashboard

The Web UI supports:

- viewing and editing runtime settings;
- adding and deleting configured model names;
- selecting one or more models;
- running a probe for one model or all selected models;
- showing success, failure, duration, attempts, and captured error output.

## Config

```json
{
  "models": ["gpt-5"],
  "timeout_seconds": 120,
  "loop_count": 1,
  "codex_command": "codex",
  "listen_addr": "0.0.0.0:8080",
  "max_output_chars": 4000
}
```

Fields:

- `models`: model names shown in the dashboard and used for selected probes.
- `timeout_seconds`: per-attempt timeout for each model.
- `loop_count`: maximum attempts per model; a model stops after the first successful attempt.
- `codex_command`: command used to invoke Codex on the VPS.
- `listen_addr`: HTTP listen address for the Web service.
- `max_output_chars`: maximum captured output returned per probe result.

## Probe Behavior

Each selected model runs in its own temporary directory. The command shape is:

```sh
codex exec --model <model> --skip-git-repo-check --ephemeral <prompt>
```

A probe succeeds only when Codex exits successfully and the captured output contains the expected short answer for the selected built-in prompt. Failures include command errors, timeouts, and expected-output mismatches. Command failure messages prefer the last captured `ERROR:` line, then the last output line, then the process error.
