package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"api-alive/internal/alive"
)

const defaultConfigPath = "config.json"

type server struct {
	configPath string
	mux        *http.ServeMux
}

type appState struct {
	Config     alive.Config `json:"config"`
	ConfigPath string       `json:"config_path"`
}

type configRequest struct {
	Models         []string `json:"models"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	LoopCount      int      `json:"loop_count"`
	CodexCommand   string   `json:"codex_command"`
	MaxOutputChars int      `json:"max_output_chars"`
	ListenAddr     string   `json:"listen_addr"`
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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	var configPath string
	fs := flag.NewFlagSet("api-alive", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&configPath, "config", defaultConfigPath, "Path to JSON config file")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected argument: %s\n", fs.Arg(0))
		return 1
	}

	srv := newServer(configPath)
	cfg, err := srv.loadConfig()
	if err != nil {
		fmt.Fprintln(stderr, "load config:", err)
		return 1
	}
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		fmt.Fprintln(stderr, "listen:", err)
		return 1
	}
	fmt.Fprintf(stdout, "api-alive listening on http://%s\n", listener.Addr().String())
	if err := http.Serve(listener, srv.mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Print(err)
		return 1
	}
	return 0
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
	writeJSON(w, appState{Config: cfg, ConfigPath: s.configPath})
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
	cfg.Models = alive.AddModels(nil, req.Models)
	cfg.TimeoutSeconds = req.TimeoutSeconds
	cfg.LoopCount = req.LoopCount
	cfg.CodexCommand = strings.TrimSpace(req.CodexCommand)
	cfg.MaxOutputChars = req.MaxOutputChars
	cfg.ListenAddr = strings.TrimSpace(req.ListenAddr)
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
	cfg.Models = models
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	runner := alive.Runner{Config: cfg, Prompts: alive.DefaultPrompts}
	ch, err := runner.Run(r.Context())
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
