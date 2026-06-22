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
	CodexCommand   string   `json:"codex_command"`
	ClaudeCommand  string   `json:"claude_command"`
	MaxOutputChars int      `json:"max_output_chars"`
}

func DefaultConfig() Config {
	return Config{
		Provider:       "codex",
		TimeoutSeconds: 120,
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

func (c *Config) ApplyDefaults() {
	if c.Provider == "" {
		c.Provider = "codex"
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 120
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
