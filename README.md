# test-api-alive

`test-api-alive` runs short liveness probes against one or more CLI-backed model names in parallel.

The default provider is Codex CLI. A Claude Code adapter is present so provider-specific command flags can evolve without changing the runner.

## Build

```sh
go build -o bin/api-alive ./cmd/api-alive
```

## Usage

Run probes for comma-separated model names:

```sh
go run ./cmd/api-alive --models gpt-5,gpt-5-mini
```

Use a JSON config file:

```sh
go run ./cmd/api-alive --config config.example.json
```

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
  "codex_command": "codex",
  "claude_command": "claude",
  "max_output_chars": 4000
}
```

CLI flags override config values for `--models`, `--provider`, and `--timeout`.

## Result Rules

Each model runs in its own temporary directory through a shell command. Human-readable results are printed as each probe finishes, with one aligned summary line per model:

```text
gpt-5              1234ms  success
gpt-5-mini          982ms  failed
```

A probe succeeds only when the provider CLI exits successfully and the captured output contains the expected short answer for the selected prompt. CLI failures, timeouts, and expected-output mismatches are reported as failures.

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
