package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"test-api-alive/internal/alive"
)

var version = "dev"

const defaultConfigPath = "config.json"

func main() {
	os.Exit(run())
}

func run() int {
	return runWithArgs(os.Args[1:], os.Stdout, os.Stderr)
}

func runWithArgs(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "list":
			return runList(args[1:], stdout, stderr)
		case "add":
			return runAdd(args[1:], stdout, stderr)
		case "remove":
			return runRemove(args[1:], stdout, stderr)
		case "exclude":
			return runExclude(args[1:], stdout, stderr)
		}
	}
	return runProbe(args, stdout, stderr)
}

func runProbe(args []string, stdout, stderr io.Writer) int {
	return runProbeWithExcludes("api-alive", args, nil, false, stdout, stderr)
}

func runExclude(args []string, stdout, stderr io.Writer) int {
	return runProbeWithExcludes("api-alive exclude", args, nil, true, stdout, stderr)
}

func runProbeWithExcludes(commandName string, args, excludePrefixes []string, requireExcludePrefixes bool, stdout, stderr io.Writer) int {
	var (
		configPath  string
		modelsCSV   string
		provider    string
		wslCommand  string
		wslDistro   string
		timeout     int
		loops       int
		jsonOutput  bool
		dryRun      bool
		listPrompts bool
		showVersion bool
	)

	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&configPath, "config", "", "Path to JSON config file (default config.json)")
	fs.StringVar(&modelsCSV, "models", "", "Comma-separated model names; overrides config.models")
	fs.StringVar(&provider, "provider", "", "Provider adapter: codex, codex-wsl, or claude")
	fs.StringVar(&wslCommand, "wsl-command", "", "WSL executable for codex-wsl provider")
	fs.StringVar(&wslDistro, "wsl-distro", "", "WSL distribution for codex-wsl provider")
	fs.IntVar(&timeout, "timeout", 0, "Per-model timeout in seconds")
	fs.IntVar(&loops, "loops", 0, "Maximum probe attempts per model")
	fs.BoolVar(&jsonOutput, "json", false, "Print one JSON object per result")
	fs.BoolVar(&dryRun, "dry-run", false, "Print commands without executing provider CLI")
	fs.BoolVar(&listPrompts, "list-prompts", false, "Print built-in prompt cases as JSON and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if requireExcludePrefixes {
		excludePrefixes = append(excludePrefixes, fs.Args()...)
		if !hasNonEmptyArg(excludePrefixes) {
			fmt.Fprintln(stderr, "at least one model prefix is required")
			return 1
		}
	} else if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unknown command or argument: %s\n", fs.Arg(0))
		return 1
	}

	if showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}
	if listPrompts {
		data, err := json.MarshalIndent(alive.DefaultPrompts, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}

	cfg, err := loadConfig(configPath, strings.TrimSpace(configPath) == "")
	if err != nil {
		fmt.Fprintln(stderr, "load config:", err)
		return 1
	}
	if models := alive.SplitCSV(modelsCSV); len(models) > 0 {
		cfg.Models = models
	}
	if strings.TrimSpace(provider) != "" {
		cfg.Provider = strings.TrimSpace(provider)
	}
	if strings.TrimSpace(wslCommand) != "" {
		cfg.WSLCommand = strings.TrimSpace(wslCommand)
	}
	if strings.TrimSpace(wslDistro) != "" {
		cfg.WSLDistro = strings.TrimSpace(wslDistro)
	}
	if timeout > 0 {
		cfg.TimeoutSeconds = timeout
	}
	if loops > 0 {
		cfg.LoopCount = loops
	}
	cfg.ApplyDefaults()
	if len(excludePrefixes) > 0 {
		cfg.Models = alive.ExcludeModelsByPrefix(cfg.Models, excludePrefixes)
	}

	providerImpl, err := alive.NewProvider(cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	runner := alive.Runner{Config: cfg, Provider: providerImpl, Prompts: alive.DefaultPrompts, DryRun: dryRun}
	results, err := runner.Run(context.Background())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	failed := false
	modelWidth := alive.ModelColumnWidth(cfg.Models)
	for res := range results {
		if jsonOutput {
			if err := alive.PrintJSONLine(stdout, res); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		} else {
			alive.PrintHumanAligned(stdout, res, modelWidth)
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

func hasNonEmptyArg(args []string) bool {
	for _, arg := range args {
		if strings.TrimSpace(arg) != "" {
			return true
		}
	}
	return false
}

func runList(args []string, stdout, stderr io.Writer) int {
	var configPath string
	fs := flag.NewFlagSet("api-alive list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&configPath, "config", "", "Path to JSON config file (default config.json)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected list argument: %s\n", fs.Arg(0))
		return 1
	}

	cfg, err := loadConfig(configPath, strings.TrimSpace(configPath) == "")
	if err != nil {
		fmt.Fprintln(stderr, "load config:", err)
		return 1
	}
	for _, model := range cfg.Models {
		fmt.Fprintln(stdout, model)
	}
	return 0
}

func runAdd(args []string, stdout, stderr io.Writer) int {
	var configPath string
	fs := flag.NewFlagSet("api-alive add", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&configPath, "config", "", "Path to JSON config file (default config.json)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, "at least one model is required")
		return 1
	}

	path := resolveConfigPath(configPath)
	cfg, err := loadConfig(configPath, true)
	if err != nil {
		fmt.Fprintln(stderr, "load config:", err)
		return 1
	}
	before := len(alive.AddModels(cfg.Models, nil))
	cfg.Models = alive.AddModels(cfg.Models, fs.Args())
	if err := alive.SaveConfig(path, cfg); err != nil {
		fmt.Fprintln(stderr, "save config:", err)
		return 1
	}
	fmt.Fprintf(stdout, "added %d model(s)\n", len(cfg.Models)-before)
	return 0
}

func runRemove(args []string, stdout, stderr io.Writer) int {
	var configPath string
	fs := flag.NewFlagSet("api-alive remove", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&configPath, "config", "", "Path to JSON config file (default config.json)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, "at least one model is required")
		return 1
	}

	path := resolveConfigPath(configPath)
	cfg, err := loadConfig(configPath, true)
	if err != nil {
		fmt.Fprintln(stderr, "load config:", err)
		return 1
	}
	before := len(alive.AddModels(cfg.Models, nil))
	cfg.Models = alive.RemoveModels(cfg.Models, fs.Args())
	if err := alive.SaveConfig(path, cfg); err != nil {
		fmt.Fprintln(stderr, "save config:", err)
		return 1
	}
	fmt.Fprintf(stdout, "removed %d model(s)\n", before-len(cfg.Models))
	return 0
}

func loadConfig(path string, allowMissing bool) (alive.Config, error) {
	cfg, err := alive.LoadConfig(resolveConfigPath(path))
	if err == nil {
		return cfg, nil
	}
	if allowMissing && errors.Is(err, os.ErrNotExist) {
		return alive.DefaultConfig(), nil
	}
	return cfg, err
}

func resolveConfigPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultConfigPath
	}
	return path
}
