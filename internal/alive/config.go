package alive

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Provider       string   `json:"provider"`
	Models         []string `json:"models"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	LoopCount      int      `json:"loop_count"`
	CodexCommand   string   `json:"codex_command"`
	ClaudeCommand  string   `json:"claude_command"`
	MaxOutputChars int      `json:"max_output_chars"`
}

func DefaultConfig() Config {
	return Config{
		Provider:       "codex",
		TimeoutSeconds: 120,
		LoopCount:      1,
		CodexCommand:   "codex",
		ClaudeCommand:  "claude",
		MaxOutputChars: 4000,
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

func SaveConfig(path string, cfg Config) error {
	cfg.ApplyDefaults()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func (c *Config) ApplyDefaults() {
	if c.Provider == "" {
		c.Provider = "codex"
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 120
	}
	if c.LoopCount <= 0 {
		c.LoopCount = 1
	}
	if c.CodexCommand == "" {
		c.CodexCommand = "codex"
	}
	if c.ClaudeCommand == "" {
		c.ClaudeCommand = "claude"
	}
	if c.MaxOutputChars <= 0 {
		c.MaxOutputChars = 4000
	}
}

func (c Config) Validate() error {
	if len(c.Models) == 0 {
		return errors.New("at least one model is required; use --models or config.models")
	}
	for _, model := range c.Models {
		if strings.TrimSpace(model) == "" {
			return errors.New("model name cannot be empty")
		}
	}
	switch c.Provider {
	case "codex", "claude":
		return nil
	default:
		return fmt.Errorf("unsupported provider %q", c.Provider)
	}
}

func (c Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

func SplitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func AddModels(existing, additions []string) []string {
	out := normalizeModels(existing)
	seen := make(map[string]struct{}, len(out)+len(additions))
	for _, model := range out {
		seen[model] = struct{}{}
	}
	for _, model := range additions {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		out = append(out, model)
		seen[model] = struct{}{}
	}
	return out
}

func RemoveModels(existing, removals []string) []string {
	remove := make(map[string]struct{}, len(removals))
	for _, model := range removals {
		model = strings.TrimSpace(model)
		if model != "" {
			remove[model] = struct{}{}
		}
	}
	out := normalizeModels(existing)
	if len(remove) == 0 {
		return out
	}
	kept := out[:0]
	for _, model := range out {
		if _, ok := remove[model]; !ok {
			kept = append(kept, model)
		}
	}
	return kept
}

func ExcludeModelsByPrefix(existing, prefixes []string) []string {
	out := normalizeModels(existing)
	prefixes = normalizePrefixes(prefixes)
	if len(prefixes) == 0 {
		return out
	}
	kept := out[:0]
	for _, model := range out {
		if !hasAnyPrefix(model, prefixes) {
			kept = append(kept, model)
		}
	}
	return kept
}

func normalizeModels(models []string) []string {
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		out = append(out, model)
		seen[model] = struct{}{}
	}
	return out
}

func normalizePrefixes(prefixes []string) []string {
	out := make([]string, 0, len(prefixes))
	seen := make(map[string]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		if _, ok := seen[prefix]; ok {
			continue
		}
		out = append(out, prefix)
		seen[prefix] = struct{}{}
	}
	return out
}

func hasAnyPrefix(model string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}
