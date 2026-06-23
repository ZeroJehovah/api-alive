# test-api-alive

`test-api-alive` runs short liveness probes against one or more CLI-backed model names in parallel.

The default provider is Codex CLI. A Claude Code adapter is present so provider-specific command flags can evolve without changing the runner.

## Build

```sh
go build -o bin/api-alive ./cmd/api-alive
```

## Usage

By default, `api-alive` reads `config.json` from the current directory.

Run probes for models from `config.json`:

```sh
go run ./cmd/api-alive
```

Run probes for comma-separated model names, overriding `config.json`:

```sh
go run ./cmd/api-alive --models gpt-5,gpt-5-mini
```

Retry each model up to three times, stopping at the first success:

```sh
go run ./cmd/api-alive --models gpt-5,gpt-5-mini --loops 3
```

Use a JSON config file:

```sh
go run ./cmd/api-alive --config config.example.json
```

List models from the current `config.json`:

```sh
go run ./cmd/api-alive list
```

Add models to the current `config.json`:

```sh
go run ./cmd/api-alive add gpt-5 gpt-5-mini
```

Remove models from the current `config.json`:

```sh
go run ./cmd/api-alive remove gpt-5-mini
```

Run probes while excluding models that match one or more prefixes for this execution only:

```sh
go run ./cmd/api-alive exclude aaa bbb
```

For example, `exclude aaa` skips configured models such as `aaa/gpt-5.5` without changing `config.json`.

Print one JSON object per completed probe:

```sh
go run ./cmd/api-alive --models gpt-5 --json
```

Preview provider commands without calling the model CLI:

```sh
go run ./cmd/api-alive --models gpt-5,gpt-5-mini --dry-run
```

List the 100 built-in prompt cases:

```sh
go run ./cmd/api-alive --list-prompts
```

## Config

```json
{
  "provider": "codex",
  "models": ["gpt-5"],
  "timeout_seconds": 120,
  "loop_count": 1,
  "codex_command": "codex",
  "claude_command": "claude",
  "max_output_chars": 4000
}
```

CLI flags override config values for `--models`, `--provider`, `--timeout`, and `--loops`.

The `list`, `add`, and `remove` commands use `config.json` by default. Pass `--config <path>` after the command to manage a different config file.

The `exclude` command runs probes like the default command, but filters the effective model list by prefix before probing. Probe flags such as `--config`, `--models`, `--provider`, `--timeout`, `--loops`, `--json`, and `--dry-run` are supported before the excluded prefixes.

## Result Rules

Each model runs in its own temporary directory through a shell command. Human-readable results are printed as each probe finishes, with one aligned summary line per model:

```text
✅ gpt-5             1.234s  attempts=1  success
❌ gpt-5-mini       0.982s  attempts=3  failed   error=ERROR: exceeded retry limit, last status: 429 Too Many Requests
```

A probe succeeds only when the provider CLI exits successfully and the captured output contains the expected short answer for the selected prompt. When `loop_count` or `--loops` is greater than 1, each model is retried until the first success or until all attempts fail. CLI failures, timeouts, and expected-output mismatches are reported as failures. Human-readable CLI failure lines prefer the last captured `ERROR:` line from the provider output, falling back to the last output line and then the process error. Human-readable failure lines include an unquoted error field truncated to 120 characters. The status field is aligned across success and failed rows; failed rows append the error detail after `failed`.

Process exit codes:

- `0`: all probes succeeded
- `1`: setup, config, or output formatting error
- `2`: at least one probe failed

## Providers

Codex command shape:

```sh
codex exec --model <model> --skip-git-repo-check --ephemeral <prompt>
```

Claude command shape:

```sh
claude --model <model> --print <prompt>
```
