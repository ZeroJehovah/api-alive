package alive

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Models         []string `json:"models"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	LoopCount      int      `json:"loop_count"`
	CodexCommand   string   `json:"codex_command"`
	ListenAddr     string   `json:"listen_addr"`
	MaxOutputChars int      `json:"max_output_chars"`
}

func DefaultConfig() Config {
	return Config{
		TimeoutSeconds: 120,
		LoopCount:      1,
		CodexCommand:   "codex",
		ListenAddr:     "0.0.0.0:8080",
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
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 120
	}
	if c.LoopCount <= 0 {
		c.LoopCount = 1
	}
	if c.CodexCommand == "" {
		c.CodexCommand = "codex"
	}
	if c.ListenAddr == "" {
		c.ListenAddr = "0.0.0.0:8080"
	}
	if c.MaxOutputChars <= 0 {
		c.MaxOutputChars = 4000
	}
}

func (c Config) Validate() error {
	if len(c.Models) == 0 {
		return errors.New("at least one model is required")
	}
	for _, model := range c.Models {
		if strings.TrimSpace(model) == "" {
			return errors.New("model name cannot be empty")
		}
	}
	return nil
}

func (c Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
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
