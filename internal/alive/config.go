package alive

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Provider        string         `json:"-"`
	Models          []string       `json:"models"`
	ModelGroups     []ModelGroup   `json:"model_groups,omitempty"`
	ModelLoopCounts map[string]int `json:"model_loop_counts,omitempty"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	// LoopCount is kept only for backward compatibility with older config files.
	LoopCount      int    `json:"loop_count,omitempty"`
	CodexCommand   string `json:"codex_command"`
	ClaudeCommand  string `json:"claude_command"`
	ListenAddr     string `json:"listen_addr"`
	MaxOutputChars int    `json:"max_output_chars"`
}

type ModelGroup struct {
	Name     string   `json:"name"`
	Provider string   `json:"provider,omitempty"`
	Models   []string `json:"models"`
}

const (
	ProviderCodex  = "codex"
	ProviderClaude = "claude"
)

func DefaultConfig() Config {
	return Config{
		Provider:       ProviderCodex,
		TimeoutSeconds: 120,
		CodexCommand:   "codex",
		ClaudeCommand:  "claude",
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
	c.Provider = normalizeProvider(c.Provider)
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = 120
	}
	if c.CodexCommand == "" {
		c.CodexCommand = "codex"
	}
	if c.ClaudeCommand == "" {
		c.ClaudeCommand = "claude"
	}
	if c.ListenAddr == "" {
		c.ListenAddr = "0.0.0.0:8080"
	}
	if c.MaxOutputChars <= 0 {
		c.MaxOutputChars = 4000
	}
	c.ModelGroups = normalizeModelGroups(c.ModelGroups, c.Models, c.Provider)
	c.Models = flattenModelGroups(c.ModelGroups)
	c.ModelLoopCounts = normalizeModelLoopCounts(c.Models, c.ModelLoopCounts, c.LoopCount)
	c.LoopCount = 0
}

func (c Config) Validate() error {
	if err := c.ValidateProviders(); err != nil {
		return err
	}
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

func (c Config) ValidateProviders() error {
	switch c.Provider {
	case ProviderCodex, ProviderClaude:
	default:
		return errors.New("unsupported provider: " + c.Provider)
	}
	for _, group := range c.ModelGroups {
		switch group.Provider {
		case ProviderCodex, ProviderClaude:
		default:
			return errors.New("unsupported group provider: " + group.Provider)
		}
	}
	return nil
}

func (c Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

func (c Config) ProviderForModel(model string) string {
	c.ApplyDefaults()
	model = strings.TrimSpace(model)
	for _, group := range c.ModelGroups {
		for _, groupModel := range group.Models {
			if groupModel == model {
				return normalizeProvider(group.Provider)
			}
		}
	}
	return normalizeProvider(c.Provider)
}

func (c Config) LoopCountForModel(model string) int {
	if c.ModelLoopCounts != nil {
		if loopCount, ok := c.ModelLoopCounts[model]; ok {
			return clampLoopCount(loopCount)
		}
	}
	if c.LoopCount > 0 {
		return clampLoopCount(c.LoopCount)
	}
	return 1
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

func AddModelsToGroup(groups []ModelGroup, additions []string, groupName string) []ModelGroup {
	return AddModelsToProviderGroup(groups, additions, groupName, ProviderCodex)
}

func AddModelsToProviderGroup(groups []ModelGroup, additions []string, groupName, provider string) []ModelGroup {
	provider = normalizeProvider(provider)
	out := normalizeModelGroups(groups, nil, provider)
	if len(out) == 0 {
		out = []ModelGroup{{Name: defaultModelGroupName(), Provider: provider}}
	}
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		groupName = defaultModelGroupName()
	}
	target := -1
	seen := make(map[string]struct{})
	for groupIndex, group := range out {
		if group.Name == groupName && target < 0 {
			target = groupIndex
		}
		for _, model := range group.Models {
			seen[model] = struct{}{}
		}
	}
	if target < 0 {
		out = append(out, ModelGroup{Name: groupName, Provider: provider})
		target = len(out) - 1
	}
	for _, model := range additions {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		out[target].Models = append(out[target].Models, model)
		seen[model] = struct{}{}
	}
	return normalizeModelGroups(out, nil, provider)
}

func RemoveModelsFromGroups(groups []ModelGroup, removals []string) []ModelGroup {
	remove := make(map[string]struct{}, len(removals))
	for _, model := range removals {
		model = strings.TrimSpace(model)
		if model != "" {
			remove[model] = struct{}{}
		}
	}
	if len(remove) == 0 {
		return normalizeModelGroups(groups, nil, ProviderCodex)
	}
	out := make([]ModelGroup, 0, len(groups))
	for _, group := range normalizeModelGroups(groups, nil, ProviderCodex) {
		kept := group.Models[:0]
		for _, model := range group.Models {
			if _, ok := remove[model]; !ok {
				kept = append(kept, model)
			}
		}
		if len(kept) > 0 {
			group.Models = kept
			out = append(out, group)
		}
	}
	return out
}

func NormalizeModelGroups(groups []ModelGroup, fallbackModels []string) []ModelGroup {
	return normalizeModelGroups(groups, fallbackModels, ProviderCodex)
}

func NormalizeModelGroupsWithProvider(groups []ModelGroup, fallbackModels []string, provider string) []ModelGroup {
	return normalizeModelGroups(groups, fallbackModels, provider)
}

func FlattenModelGroups(groups []ModelGroup) []string {
	return flattenModelGroups(groups)
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

func normalizeModelGroups(groups []ModelGroup, fallbackModels []string, defaultProvider string) []ModelGroup {
	defaultProvider = normalizeProvider(defaultProvider)
	seen := make(map[string]struct{})
	out := make([]ModelGroup, 0, len(groups))
	for _, group := range groups {
		name := strings.TrimSpace(group.Name)
		if name == "" {
			name = defaultModelGroupName()
		}
		provider := normalizeProvider(group.Provider)
		if group.Provider == "" {
			provider = defaultProvider
		}
		models := make([]string, 0, len(group.Models))
		for _, model := range group.Models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			if _, ok := seen[model]; ok {
				continue
			}
			models = append(models, model)
			seen[model] = struct{}{}
		}
		if len(models) == 0 {
			continue
		}
		out = append(out, ModelGroup{Name: name, Provider: provider, Models: models})
	}
	if len(out) == 0 {
		models := normalizeModels(fallbackModels)
		if len(models) > 0 {
			out = append(out, ModelGroup{Name: defaultModelGroupName(), Provider: defaultProvider, Models: models})
		}
	}
	return out
}

func defaultModelGroupName() string {
	return "Default"
}

func flattenModelGroups(groups []ModelGroup) []string {
	models := make([]string, 0)
	for _, group := range groups {
		models = append(models, group.Models...)
	}
	return normalizeModels(models)
}

func normalizeModelLoopCounts(models []string, counts map[string]int, legacyLoopCount int) map[string]int {
	out := make(map[string]int, len(models))
	defaultLoopCount := legacyLoopCount
	if defaultLoopCount <= 0 {
		defaultLoopCount = 1
	}
	for _, model := range models {
		loopCount := defaultLoopCount
		if counts != nil {
			if configured, ok := counts[model]; ok {
				loopCount = configured
			}
		}
		out[model] = clampLoopCount(loopCount)
	}
	return out
}

func clampLoopCount(loopCount int) int {
	if loopCount < 0 {
		return 1
	}
	if loopCount > 99 {
		return 99
	}
	return loopCount
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return ProviderCodex
	}
	return provider
}
