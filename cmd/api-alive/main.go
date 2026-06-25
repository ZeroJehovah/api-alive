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
	Models          []string       `json:"models"`
	ModelLoopCounts map[string]int `json:"model_loop_counts"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	CodexCommand    string         `json:"codex_command"`
	MaxOutputChars  int            `json:"max_output_chars"`
	ListenAddr      string         `json:"listen_addr"`
}

type modelsRequest struct {
	Models []string `json:"models"`
}

type probeRequest struct {
	Models          []string       `json:"models"`
	ModelLoopCounts map[string]int `json:"model_loop_counts"`
	Stream          bool           `json:"stream"`
}

type stopProbeRequest struct {
	Model string `json:"model"`
}

type probeResponse struct {
	Task   probeTask    `json:"task"`
	Config alive.Config `json:"config,omitempty"`
}

type probeTask struct {
	ID            int64           `json:"id"`
	Running       bool            `json:"running"`
	Models        []string        `json:"models"`
	RunningModels []string        `json:"running_models"`
	Results       []alive.Result  `json:"results"`
	Logs          []probeLogEntry `json:"logs"`
	StartedAt     string          `json:"started_at,omitempty"`
	FinishedAt    string          `json:"finished_at,omitempty"`
	Error         string          `json:"error,omitempty"`
	LoopCounts    map[string]int  `json:"loop_counts,omitempty"`
}

type probeLogEntry struct {
	Time      string       `json:"time"`
	LoopCount int          `json:"loop_count"`
	Result    alive.Result `json:"result"`
}

type taskStore struct {
	mu        sync.Mutex
	nextID    int64
	nextRunID int64
	task      probeTask
	runs      map[string]activeModelRun
}

type activeModelRun struct {
	ID     int64
	Model  string
	Cancel context.CancelFunc
}

type modelRun struct {
	ID    int64
	Model string
	Ctx   context.Context
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
	s.mux.HandleFunc("/api/probe/results/clear", s.handleClearProbeResults)
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
	cfg.ModelLoopCounts = req.ModelLoopCounts
	cfg.TimeoutSeconds = req.TimeoutSeconds
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
	for _, model := range models {
		if cfg.ModelLoopCounts == nil {
			cfg.ModelLoopCounts = make(map[string]int)
		}
		if count, ok := req.ModelLoopCounts[model]; ok {
			cfg.ModelLoopCounts[model] = count
		}
	}
	cfg.ApplyDefaults()
	if err := alive.SaveConfig(s.configPath, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	runnerCfg := cfg
	runnerCfg.Models = models
	runnerCfg.ApplyDefaults()
	if err := runnerCfg.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	task, runs, err := s.tasks.start(models, loopCountsForModels(models, runnerCfg.ModelLoopCounts))
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	for _, run := range runs {
		modelCfg := runnerCfg
		modelCfg.Models = []string{run.Model}
		runner := alive.Runner{Config: modelCfg, Prompts: alive.DefaultPrompts}
		go s.runProbeTask(run.Ctx, task.ID, run.ID, runner)
	}
	writeJSON(w, probeResponse{Task: task, Config: cfg})
}

func (s *server) handleStopProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req stopProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	models := alive.AddModels(nil, []string{req.Model})
	if len(models) == 0 {
		writeError(w, http.StatusBadRequest, "select at least one running model to stop")
		return
	}
	task, err := s.tasks.stopModels(models)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, probeResponse{Task: task})
}

func (s *server) handleClearProbeResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, probeResponse{Task: s.tasks.clearResults()})
}

func (s *server) runProbeTask(ctx context.Context, taskID, runID int64, runner alive.Runner) {
	events, err := runner.RunEvents(ctx)
	model := ""
	if len(runner.Config.Models) > 0 {
		model = runner.Config.Models[0]
	}
	if err != nil {
		s.tasks.failRun(taskID, runID, model, err.Error())
		return
	}
	for event := range events {
		s.tasks.applyEvent(taskID, runID, event)
	}
	s.tasks.finishRun(taskID, runID, model)
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

func (ts *taskStore) start(models []string, loopCounts map[string]int) (probeTask, []modelRun, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.runs == nil {
		ts.runs = make(map[string]activeModelRun)
	}

	if !ts.task.Running {
		logs := append([]probeLogEntry(nil), ts.task.Logs...)
		results := append([]alive.Result(nil), ts.task.Results...)
		for _, res := range initialRunningResults(models) {
			results = upsertResult(results, res)
		}
		ts.nextID++
		now := time.Now().Format(time.RFC3339)
		ts.task = probeTask{
			ID:            ts.nextID,
			Running:       true,
			Models:        append([]string(nil), models...),
			RunningModels: append([]string(nil), models...),
			StartedAt:     now,
			Results:       results,
			Logs:          logs,
			LoopCounts:    loopCountsForModels(models, loopCounts),
		}
		ts.runs = make(map[string]activeModelRun)
	} else {
		ts.task.Models = alive.AddModels(ts.task.Models, models)
		for _, res := range initialRunningResults(models) {
			ts.task.Results = upsertResult(ts.task.Results, res)
		}
		if ts.task.LoopCounts == nil {
			ts.task.LoopCounts = make(map[string]int)
		}
		for _, model := range models {
			ts.task.LoopCounts[model] = loopCounts[model]
		}
		ts.task.FinishedAt = ""
		ts.task.Error = ""
	}

	runs := make([]modelRun, 0, len(models))
	for _, model := range models {
		if existing, ok := ts.runs[model]; ok {
			existing.Cancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		ts.nextRunID++
		run := activeModelRun{ID: ts.nextRunID, Model: model, Cancel: cancel}
		ts.runs[model] = run
		runs = append(runs, modelRun{ID: run.ID, Model: model, Ctx: ctx})
	}
	ts.task.RunningModels = runningModelsInOrder(ts.task.Models, ts.runs)
	return cloneTask(ts.task), runs, nil
}

func (ts *taskStore) stopModels(models []string) (probeTask, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.task.Running {
		return cloneTask(ts.task), errors.New("no probe task is running")
	}
	stopped := 0
	for _, model := range models {
		run, ok := ts.runs[model]
		if !ok {
			continue
		}
		run.Cancel()
		stopped++
	}
	if stopped == 0 {
		return cloneTask(ts.task), errors.New("none of the selected models are running")
	}
	return cloneTask(ts.task), nil
}

func (ts *taskStore) clearResults() probeTask {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.task.Results) == 0 {
		return cloneTask(ts.task)
	}
	running := make(map[string]struct{}, len(ts.task.RunningModels))
	for _, model := range ts.task.RunningModels {
		running[model] = struct{}{}
	}
	kept := ts.task.Results[:0]
	for _, res := range ts.task.Results {
		if _, ok := running[res.Model]; ok {
			kept = append(kept, res)
		}
	}
	ts.task.Results = kept
	return cloneTask(ts.task)
}

func (ts *taskStore) snapshot() probeTask {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return cloneTask(ts.task)
}

func (ts *taskStore) applyEvent(taskID, runID int64, event alive.Event) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID {
		return
	}
	res := event.Result
	if !ts.isCurrentRunLocked(runID, res.Model) {
		return
	}
	now := time.Now().Format(time.RFC3339)
	if event.Type == alive.EventAttempt {
		ts.task.Results = upsertResult(ts.task.Results, timestampResult(displayAttemptResult(res, ts.loopCountForModel(res.Model)), now))
		ts.prependLogLocked(now, res)
		return
	}
	if event.Type == alive.EventResult {
		ts.task.Results = upsertResult(ts.task.Results, timestampResult(res, now))
		ts.task.RunningModels = removeModelName(ts.task.RunningModels, res.Model)
		if len(res.AttemptResults) == 0 {
			ts.prependLogLocked(now, res)
		}
	}
}

func (ts *taskStore) finishRun(taskID, runID int64, model string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID || !ts.isCurrentRunLocked(runID, model) {
		return
	}
	delete(ts.runs, model)
	ts.task.RunningModels = removeModelName(ts.task.RunningModels, model)
	if len(ts.runs) > 0 {
		return
	}
	ts.task.Running = false
	ts.task.RunningModels = nil
	ts.task.FinishedAt = time.Now().Format(time.RFC3339)
	ts.runs = nil
}

func (ts *taskStore) failRun(taskID, runID int64, model, message string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.task.ID != taskID || !ts.isCurrentRunLocked(runID, model) {
		return
	}
	delete(ts.runs, model)
	ts.task.RunningModels = removeModelName(ts.task.RunningModels, model)
	ts.task.Error = message
	if len(ts.runs) > 0 {
		return
	}
	ts.task.Running = false
	ts.task.RunningModels = nil
	ts.task.FinishedAt = time.Now().Format(time.RFC3339)
	ts.runs = nil
}

func (ts *taskStore) isCurrentRunLocked(runID int64, model string) bool {
	run, ok := ts.runs[model]
	return ok && run.ID == runID
}

func (ts *taskStore) loopCountForModel(model string) int {
	if ts.task.LoopCounts != nil {
		if loopCount, ok := ts.task.LoopCounts[model]; ok {
			return loopCount
		}
	}
	return 1
}

func (ts *taskStore) prependLogLocked(timestamp string, res alive.Result) {
	entry := probeLogEntry{Time: timestamp, LoopCount: ts.loopCountForModel(res.Model), Result: res}
	ts.task.Logs = append([]probeLogEntry{entry}, ts.task.Logs...)
	if len(ts.task.Logs) > maxServerLogEntries {
		ts.task.Logs = ts.task.Logs[:maxServerLogEntries]
	}
}

func timestampResult(res alive.Result, timestamp string) alive.Result {
	res.UpdatedAt = timestamp
	return res
}

func initialRunningResults(models []string) []alive.Result {
	results := make([]alive.Result, 0, len(models))
	for _, model := range models {
		results = append(results, alive.Result{Model: model, Attempts: 1})
	}
	return results
}

func loopCountsForModels(models []string, counts map[string]int) map[string]int {
	loopCounts := make(map[string]int, len(models))
	for _, model := range models {
		loopCount := 1
		if counts != nil {
			if configured, ok := counts[model]; ok {
				loopCount = configured
			}
		}
		if loopCount < 0 {
			loopCount = 1
		}
		if loopCount > 99 {
			loopCount = 99
		}
		loopCounts[model] = loopCount
	}
	return loopCounts
}

func runningModelsInOrder(models []string, runs map[string]activeModelRun) []string {
	running := make([]string, 0, len(runs))
	seen := make(map[string]struct{}, len(runs))
	for _, model := range models {
		if _, ok := runs[model]; ok {
			if _, exists := seen[model]; !exists {
				running = append(running, model)
				seen[model] = struct{}{}
			}
		}
	}
	for model := range runs {
		if _, exists := seen[model]; !exists {
			running = append(running, model)
		}
	}
	return running
}

func displayAttemptResult(res alive.Result, loopCount int) alive.Result {
	if !res.Success && (loopCount == 0 || res.Attempts < loopCount) {
		res.Attempts++
	}
	return res
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
	if task.LoopCounts != nil {
		loopCounts := make(map[string]int, len(task.LoopCounts))
		for model, loopCount := range task.LoopCounts {
			loopCounts[model] = loopCount
		}
		task.LoopCounts = loopCounts
	}
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
