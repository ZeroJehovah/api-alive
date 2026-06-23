package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf16"

	"test-api-alive/internal/alive"
)

const defaultConfigFile = "config.json"

type server struct {
	configPath string
	mux        *http.ServeMux
}

type appState struct {
	Config       alive.Config `json:"config"`
	WSLDistros   []string     `json:"wsl_distros"`
	ConfigPath   string       `json:"config_path"`
	DefaultModel string       `json:"default_model"`
}

type configRequest struct {
	Provider       string   `json:"provider"`
	Models         []string `json:"models"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	LoopCount      int      `json:"loop_count"`
	CodexCommand   string   `json:"codex_command"`
	ClaudeCommand  string   `json:"claude_command"`
	WSLCommand     string   `json:"wsl_command"`
	WSLDistro      string   `json:"wsl_distro"`
	MaxOutputChars int      `json:"max_output_chars"`
}

type modelsRequest struct {
	Models []string `json:"models"`
}

type probeRequest struct {
	Models []string `json:"models"`
}

type probeResponse struct {
	Results []alive.Result `json:"results"`
}

func main() {
	configPath := defaultConfigFile
	if len(os.Args) > 1 {
		configPath = strings.TrimSpace(os.Args[1])
	}
	if configPath == "" {
		configPath = defaultConfigFile
	}

	srv := newServer(configPath)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	url := "http://" + listener.Addr().String()
	openBrowser(url)
	log.Printf("api-alive gui: %s", url)
	if err := http.Serve(listener, srv.mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newServer(configPath string) *server {
	s := &server{configPath: configPath, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/state", s.handleState)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/models", s.handleModels)
	s.mux.HandleFunc("/api/probe", s.handleProbe)
	return s
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexHTML)
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cfg, err := s.loadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	distros, _ := listWSLDistros(cfg.WSLCommand)
	writeJSON(w, appState{
		Config:       cfg,
		WSLDistros:   distros,
		ConfigPath:   s.configPath,
		DefaultModel: firstModel(cfg.Models),
	})
}

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	cfg, err := s.loadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg.Provider = strings.TrimSpace(req.Provider)
	cfg.Models = alive.AddModels(nil, req.Models)
	cfg.TimeoutSeconds = req.TimeoutSeconds
	cfg.LoopCount = req.LoopCount
	cfg.CodexCommand = strings.TrimSpace(req.CodexCommand)
	cfg.ClaudeCommand = strings.TrimSpace(req.ClaudeCommand)
	cfg.WSLCommand = strings.TrimSpace(req.WSLCommand)
	cfg.WSLDistro = strings.TrimSpace(req.WSLDistro)
	cfg.MaxOutputChars = req.MaxOutputChars
	cfg.ApplyDefaults()
	if err := alive.SaveConfig(s.configPath, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, cfg)
}

func (s *server) handleModels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req modelsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		cfg, err := s.loadConfig()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cfg.Models = alive.AddModels(cfg.Models, req.Models)
		if err := alive.SaveConfig(s.configPath, cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, cfg)
	case http.MethodDelete:
		var req modelsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		cfg, err := s.loadConfig()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cfg.Models = alive.RemoveModels(cfg.Models, req.Models)
		if err := alive.SaveConfig(s.configPath, cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, cfg)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req probeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	models := alive.AddModels(nil, req.Models)
	if len(models) == 0 {
		writeError(w, http.StatusBadRequest, "select at least one model")
		return
	}
	cfg, err := s.loadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg.Provider = "codex-wsl"
	cfg.Models = models
	cfg.ApplyDefaults()
	provider, err := alive.NewProvider(cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	runner := alive.Runner{Config: cfg, Provider: provider, Prompts: alive.DefaultPrompts}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	ch, err := runner.Run(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	results := make([]alive.Result, 0, len(models))
	for result := range ch {
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Model < results[j].Model
	})
	writeJSON(w, probeResponse{Results: results})
}

func (s *server) loadConfig() (alive.Config, error) {
	cfg, err := alive.LoadConfig(s.configPath)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		cfg = alive.DefaultConfig()
		cfg.Provider = "codex-wsl"
		cfg.ApplyDefaults()
		if err := ensureConfigDir(s.configPath); err != nil {
			return cfg, err
		}
		if err := alive.SaveConfig(s.configPath, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	return cfg, err
}

func ensureConfigDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func listWSLDistros(wslCommand string) ([]string, error) {
	command := strings.TrimSpace(wslCommand)
	if command == "" {
		command = "wsl.exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, command, "-l", "-q").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(decodeWSLOutput(output), "\r\n", "\n"), "\n")
	distros := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Trim(strings.TrimSpace(line), "\x00")
		if line != "" {
			distros = append(distros, line)
		}
	}
	return distros, nil
}

func decodeWSLOutput(output []byte) string {
	if len(output) < 2 || !looksLikeUTF16LE(output) {
		return string(output)
	}
	u16 := make([]uint16, 0, len(output)/2)
	for i := 0; i+1 < len(output); i += 2 {
		u16 = append(u16, uint16(output[i])|uint16(output[i+1])<<8)
	}
	return string(utf16.Decode(u16))
}

func looksLikeUTF16LE(output []byte) bool {
	zeros := 0
	for i := 1; i < len(output); i += 2 {
		if output[i] == 0 {
			zeros++
		}
	}
	return zeros > len(output)/4
}

func firstModel(models []string) string {
	if len(models) == 0 {
		return ""
	}
	return models[0]
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
