package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"test-api-alive/internal/alive"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		configPath  string
		modelsCSV   string
		provider    string
		timeout     int
		jsonOutput  bool
		dryRun      bool
		listPrompts bool
		showVersion bool
	)

	flag.StringVar(&configPath, "config", "", "Path to JSON config file")
	flag.StringVar(&modelsCSV, "models", "", "Comma-separated model names; overrides config.models")
	flag.StringVar(&provider, "provider", "", "Provider adapter: codex or claude")
	flag.IntVar(&timeout, "timeout", 0, "Per-model timeout in seconds")
	flag.BoolVar(&jsonOutput, "json", false, "Print one JSON object per result")
	flag.BoolVar(&dryRun, "dry-run", false, "Print commands without executing provider CLI")
	flag.BoolVar(&listPrompts, "list-prompts", false, "Print built-in prompt cases as JSON and exit")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return 0
	}
	if listPrompts {
		data, err := json.MarshalIndent(alive.DefaultPrompts, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(string(data))
		return 0
	}

	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		return 1
	}
	if models := alive.SplitCSV(modelsCSV); len(models) > 0 {
		cfg.Models = models
	}
	if strings.TrimSpace(provider) != "" {
		cfg.Provider = strings.TrimSpace(provider)
	}
	if timeout > 0 {
		cfg.TimeoutSeconds = timeout
	}
	cfg.ApplyDefaults()

	providerImpl, err := alive.NewProvider(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	runner := alive.Runner{Config: cfg, Provider: providerImpl, Prompts: alive.DefaultPrompts, DryRun: dryRun}
	results, err := runner.Run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	failed := false
	modelWidth := alive.ModelColumnWidth(cfg.Models)
	for res := range results {
		if jsonOutput {
			if err := alive.PrintJSONLine(os.Stdout, res); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		} else {
			alive.PrintHumanAligned(os.Stdout, res, modelWidth)
		}
		if !res.Success {
			failed = true
		}
	}
	if failed {
		return 2
	}
	return 0
}
