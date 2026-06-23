package main

import (
	"context"
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
	"strings"
	"sync"
	"time"

	"api-alive/internal/alive"
)

const defaultConfigPath = "config.json"

type server struct {
	configPath string
	mux        *http.ServeMux
	tasks      *taskStore
}

type appState struct {
	Config     alive.Config `json:"config"`
	ConfigPath string       `json:"config_path"`
	Task       probeTask    `json:"task"`
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
	Stream bool     `json:"stream"`
}

type probeResponse struct {
	Task probeTask `json:"task"`
}

type probeTask struct {
	ID            int64           `json:"id"`
	Running       bool            `json:"running"`
	Stopping      bool            `json:"stopping"`
	Models        []string        `json:"models"`
	RunningModels []string        `json:"running_models"`
	Results       []alive.Result  `json:"results"`
	Logs          []probeLogEntry `json:"logs"`
	StartedAt     string          `json:"started_at,omitempty"`
	FinishedAt    string          `json:"finished_at,omitempty"`
	Error         string          `json:"error,omitempty"`
}

type probeLogEntry struct {
	Time   string       `json:"time"`
	Result alive.Result `json:"result"`
}

type taskStore struct {
	mu     sync.Mutex
	nextID int64
	task   probeTask
	cancel context.CancelFunc
}

const maxServerLogEntries = 100

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
	s := &server{configPath: configPath, mux: http.NewServeMux(), tasks: &taskStore{}}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/state", s.handleState)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/models", s.handleModels)
	s.mux.HandleFunc("/api/probe", s.handleProbe)
	s.mux.HandleFunc("/api/probe/stop", s.handleStopProbe)
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
	writeJSON(w, appState{Config: cfg, ConfigPath: s.configPath, Task: s.tasks.snapshot()})
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
	task, ctx, err := s.tasks.start(models)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	go s.runProbeTask(ctx, task.ID, runner)
	writeJSON(w, probeResponse{Task: task})
}

func (s *server) handleStopProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	task, err := s.tasks.stop()
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, probeResponse{Task: task})
}

func (s *server) runProbeTask(ctx context.Context, taskID int64, runner alive.Runner) {
	events, err := runner.RunEvents(ctx)
	if err != nil {
		s.tasks.fail(taskID, err.Error())
		return
	}
	for event := range events {
		s.tasks.applyEvent(taskID, event)
	}
	s.tasks.finish(taskID)
}

func (s *server) streamProbe(w http.ResponseWriter, r *http.Request, runner alive.Runner) {
	events, err := runner.RunEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)
	encoder := json.NewEncoder(w)
	for event := range events {
		if err := encoder.Encode(event); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}

func (ts *taskStore) start(models []string) (probeTask, context.Context, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.Running {
		return probeTask{}, nil, errors.New("a probe task is already running")
	}
	logs := append([]probeLogEntry(nil), ts.task.Logs...)
	ts.nextID++
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now().Format(time.RFC3339)
	task := probeTask{
		ID:            ts.nextID,
		Running:       true,
		Models:        append([]string(nil), models...),
		RunningModels: append([]string(nil), models...),
		StartedAt:     now,
		Results:       []alive.Result{},
		Logs:          logs,
	}
	ts.task = task
	ts.cancel = cancel
	return cloneTask(ts.task), ctx, nil
}

func (ts *taskStore) stop() (probeTask, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.task.Running {
		return cloneTask(ts.task), errors.New("no probe task is running")
	}
	ts.task.Stopping = true
	if ts.cancel != nil {
		ts.cancel()
	}
	return cloneTask(ts.task), nil
}

func (ts *taskStore) snapshot() probeTask {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return cloneTask(ts.task)
}

func (ts *taskStore) applyEvent(taskID int64, event alive.Event) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID {
		return
	}
	res := event.Result
	if event.Type == alive.EventAttempt {
		ts.prependLogLocked(res)
		return
	}
	if event.Type == alive.EventResult {
		ts.task.Results = upsertResult(ts.task.Results, res)
		ts.task.RunningModels = removeModelName(ts.task.RunningModels, res.Model)
		if len(res.AttemptResults) == 0 {
			ts.prependLogLocked(res)
		}
	}
}

func (ts *taskStore) finish(taskID int64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID {
		return
	}
	ts.task.Running = false
	ts.task.Stopping = false
	ts.task.RunningModels = nil
	ts.task.FinishedAt = time.Now().Format(time.RFC3339)
	ts.cancel = nil
}

func (ts *taskStore) fail(taskID int64, message string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID {
		return
	}
	ts.task.Running = false
	ts.task.Stopping = false
	ts.task.RunningModels = nil
	ts.task.Error = message
	ts.task.FinishedAt = time.Now().Format(time.RFC3339)
	ts.cancel = nil
}

func (ts *taskStore) prependLogLocked(res alive.Result) {
	entry := probeLogEntry{Time: time.Now().Format(time.RFC3339), Result: res}
	ts.task.Logs = append([]probeLogEntry{entry}, ts.task.Logs...)
	if len(ts.task.Logs) > maxServerLogEntries {
		ts.task.Logs = ts.task.Logs[:maxServerLogEntries]
	}
}

func upsertResult(results []alive.Result, res alive.Result) []alive.Result {
	for i := range results {
		if results[i].Model == res.Model {
			results[i] = res
			return results
		}
	}
	return append(results, res)
}

func removeModelName(models []string, model string) []string {
	kept := models[:0]
	for _, existing := range models {
		if existing != model {
			kept = append(kept, existing)
		}
	}
	return kept
}

func cloneTask(task probeTask) probeTask {
	task.Models = append([]string(nil), task.Models...)
	task.RunningModels = append([]string(nil), task.RunningModels...)
	task.Results = append([]alive.Result(nil), task.Results...)
	task.Logs = append([]probeLogEntry(nil), task.Logs...)
	return task
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
