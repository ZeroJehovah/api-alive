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
	Models          []string           `json:"models"`
	ModelGroups     []alive.ModelGroup `json:"model_groups"`
	ModelLoopCounts map[string]int     `json:"model_loop_counts"`
	TimeoutSeconds  int                `json:"timeout_seconds"`
	CodexCommand    string             `json:"codex_command"`
	ClaudeCommand   string             `json:"claude_command"`
	MaxOutputChars  int                `json:"max_output_chars"`
	ListenAddr      string             `json:"listen_addr"`
}

type modelsRequest struct {
	Models   []string `json:"models"`
	Group    string   `json:"group"`
	Provider string   `json:"provider"`
}

type probeRequest struct {
	Models          []string       `json:"models"`
	ModelLoopCounts map[string]int `json:"model_loop_counts"`
	Stream          bool           `json:"stream"`
}

type stopProbeRequest struct {
	Model  string   `json:"model"`
	Models []string `json:"models"`
}

type probeResponse struct {
	Task   probeTask    `json:"task"`
	Config alive.Config `json:"config,omitempty"`
}

type probeTask struct {
	ID            int64           `json:"id"`
	Revision      int64           `json:"revision"`
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
	mu          sync.Mutex
	nextID      int64
	nextRunID   int64
	task        probeTask
	runs        map[string]activeModelRun
	subscribers map[chan probeTask]struct{}
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
const runLivenessCheckInterval = 500 * time.Millisecond

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
	s.mux.HandleFunc("/api/events", s.handleEvents)
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
	cfg.ModelGroups = req.ModelGroups
	cfg.Models = req.Models
	cfg.ModelLoopCounts = req.ModelLoopCounts
	cfg.TimeoutSeconds = req.TimeoutSeconds
	cfg.CodexCommand = strings.TrimSpace(req.CodexCommand)
	cfg.ClaudeCommand = strings.TrimSpace(req.ClaudeCommand)
	cfg.MaxOutputChars = req.MaxOutputChars
	cfg.ListenAddr = strings.TrimSpace(req.ListenAddr)
	cfg.ApplyDefaults()
	if err := cfg.ValidateProviders(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
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
		provider := strings.TrimSpace(req.Provider)
		if provider == "" {
			provider = cfg.Provider
		}
		cfg.ModelGroups = alive.AddModelsToProviderGroup(cfg.ModelGroups, req.Models, req.Group, provider)
		cfg.Models = alive.FlattenModelGroups(cfg.ModelGroups)
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
		cfg.ModelGroups = alive.RemoveModelsFromGroups(cfg.ModelGroups, req.Models)
		cfg.Models = alive.FlattenModelGroups(cfg.ModelGroups)
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
	runnerCfg := probeRunnerConfig(cfg, models)
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
		runner := alive.Runner{Config: singleModelRunnerConfig(runnerCfg, run.Model), Prompts: alive.DefaultPrompts}
		go s.runProbeTask(run.Ctx, task.ID, run.ID, runner)
	}
	writeJSON(w, probeResponse{Task: task, Config: cfg})
}

func probeRunnerConfig(cfg alive.Config, models []string) alive.Config {
	cfg.ApplyDefaults()
	runnerCfg := cfg
	runnerCfg.Models = append([]string(nil), models...)
	runnerCfg.ModelGroups = modelGroupsForModels(cfg.ModelGroups, models, cfg.Provider)
	runnerCfg.ApplyDefaults()
	return runnerCfg
}

func singleModelRunnerConfig(cfg alive.Config, model string) alive.Config {
	modelCfg := cfg
	modelCfg.Provider = cfg.ProviderForModel(model)
	modelCfg.Models = []string{model}
	modelCfg.ModelGroups = nil
	modelCfg.ApplyDefaults()
	return modelCfg
}

func modelGroupsForModels(groups []alive.ModelGroup, models []string, defaultProvider string) []alive.ModelGroup {
	wanted := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model != "" {
			wanted[model] = struct{}{}
		}
	}
	selected := make([]alive.ModelGroup, 0, len(groups))
	matched := make(map[string]struct{}, len(wanted))
	for _, group := range groups {
		kept := make([]string, 0, len(group.Models))
		for _, model := range group.Models {
			if _, ok := wanted[model]; ok {
				kept = append(kept, model)
				matched[model] = struct{}{}
			}
		}
		if len(kept) > 0 {
			selected = append(selected, alive.ModelGroup{Name: group.Name, Provider: group.Provider, Models: kept})
		}
	}
	unmatched := make([]string, 0)
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := matched[model]; !ok {
			unmatched = append(unmatched, model)
		}
	}
	if len(unmatched) > 0 {
		selected = append(selected, alive.ModelGroup{Name: "Default", Provider: defaultProvider, Models: unmatched})
	}
	return selected
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
	models := alive.AddModels(nil, req.Models)
	models = alive.AddModels(models, []string{req.Model})
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

func (s *server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	updates, unsubscribe := s.tasks.subscribe()
	defer unsubscribe()

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	writeEvent := func(task probeTask) error {
		data, err := json.Marshal(task)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case task, ok := <-updates:
			if !ok {
				return
			}
			if err := writeEvent(task); err != nil {
				return
			}
		case <-ping.C:
			if _, err := io.WriteString(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *server) runProbeTask(ctx context.Context, taskID, runID int64, runner alive.Runner) {
	model := ""
	if len(runner.Config.Models) > 0 {
		model = runner.Config.Models[0]
	}
	ctx, cancel := s.currentRunContext(ctx, taskID, runID, model)
	defer cancel()
	events, err := runner.RunEvents(ctx)
	if err != nil {
		s.tasks.failRun(taskID, runID, model, err.Error())
		return
	}
	for event := range events {
		if !s.tasks.applyEvent(taskID, runID, event) {
			cancel()
		}
	}
	s.tasks.finishRun(taskID, runID, model)
}

func (s *server) currentRunContext(parent context.Context, taskID, runID int64, model string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	if model == "" {
		return ctx, cancel
	}
	if !s.tasks.runIsCurrent(taskID, runID, model) {
		cancel()
		return ctx, cancel
	}
	go func() {
		ticker := time.NewTicker(runLivenessCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !s.tasks.runIsCurrent(taskID, runID, model) {
					cancel()
					return
				}
			}
		}
	}()
	return ctx, cancel
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

	if len(ts.runs) == 0 {
		ts.nextID++
		now := time.Now().Format(time.RFC3339)
		revision := ts.task.Revision
		results := append([]alive.Result(nil), ts.task.Results...)
		for _, res := range initialRunningResults(models) {
			results = upsertResult(results, res)
		}
		mergedLoopCounts := make(map[string]int, len(ts.task.LoopCounts)+len(models))
		for model, loopCount := range ts.task.LoopCounts {
			mergedLoopCounts[model] = loopCount
		}
		for _, model := range models {
			mergedLoopCounts[model] = loopCounts[model]
		}
		ts.task = probeTask{
			ID:            ts.nextID,
			Revision:      revision,
			Running:       true,
			Models:        alive.AddModels(ts.task.Models, models),
			RunningModels: append([]string(nil), models...),
			StartedAt:     now,
			Results:       results,
			Logs:          append([]probeLogEntry(nil), ts.task.Logs...),
			LoopCounts:    mergedLoopCounts,
		}
		ts.runs = make(map[string]activeModelRun)
	} else {
		ts.task.Running = true
		ts.task.FinishedAt = ""
		ts.task.Models = alive.AddModels(ts.task.Models, activeRunModels(ts.runs))
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
	ts.task.Running = len(ts.runs) > 0
	ts.publishLocked()
	return cloneTask(ts.task), runs, nil
}

func (ts *taskStore) stopModels(models []string) (probeTask, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.runs) == 0 {
		ts.task.Running = false
		ts.task.RunningModels = nil
		return cloneTask(ts.task), errors.New("no probe task is running")
	}
	stopped := 0
	now := time.Now().Format(time.RFC3339)
	for _, model := range models {
		run, ok := ts.runs[model]
		if !ok {
			continue
		}
		run.Cancel()
		delete(ts.runs, model)
		res := resultByModelName(ts.task.Results, model)
		if res.Attempts < 1 {
			res.Attempts = 1
		}
		res.Model = model
		res.Success = false
		res.ExitCode = nil
		res.Error = context.Canceled.Error()
		res.Duration = 0
		res.DurationMS = 0
		res.UpdatedAt = now
		res.AttemptResults = nil
		ts.task.Results = upsertResult(ts.task.Results, res)
		ts.prependLogLocked(now, res)
		stopped++
	}
	if stopped == 0 {
		return cloneTask(ts.task), errors.New("none of the selected models are running")
	}
	ts.task.RunningModels = runningModelsInOrder(ts.task.Models, ts.runs)
	ts.task.Running = len(ts.runs) > 0
	if !ts.task.Running {
		ts.task.RunningModels = nil
		ts.task.FinishedAt = now
		ts.runs = nil
	}
	ts.publishLocked()
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
	ts.publishLocked()
	return cloneTask(ts.task)
}

func (ts *taskStore) snapshot() probeTask {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.reconcileRunningLocked()
	return cloneTask(ts.task)
}

func (ts *taskStore) subscribe() (<-chan probeTask, func()) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.reconcileRunningLocked()
	if ts.subscribers == nil {
		ts.subscribers = make(map[chan probeTask]struct{})
	}
	ch := make(chan probeTask, 1)
	ts.subscribers[ch] = struct{}{}
	ch <- cloneTask(ts.task)
	unsubscribe := func() {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		if _, ok := ts.subscribers[ch]; !ok {
			return
		}
		delete(ts.subscribers, ch)
		close(ch)
	}
	return ch, unsubscribe
}

func (ts *taskStore) publishLocked() {
	ts.task.Revision++
	if len(ts.subscribers) == 0 {
		return
	}
	snapshot := cloneTask(ts.task)
	for ch := range ts.subscribers {
		select {
		case ch <- snapshot:
			continue
		default:
		}
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- snapshot:
		default:
		}
	}
}

func (ts *taskStore) applyEvent(taskID, runID int64, event alive.Event) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	res := event.Result
	if !ts.isCurrentRunLocked(runID, res.Model) {
		return false
	}
	if ts.task.ID != taskID {
		if event.Type == alive.EventResult {
			if ts.completeRunLocked(runID, res.Model, time.Now().Format(time.RFC3339)) {
				ts.publishLocked()
			}
		}
		return false
	}
	now := time.Now().Format(time.RFC3339)
	if event.Type == alive.EventAttempt {
		ts.task.Results = upsertResult(ts.task.Results, timestampResult(displayAttemptResult(res, ts.loopCountForModel(res.Model)), now))
		ts.prependLogLocked(now, res)
		ts.publishLocked()
		return true
	}
	if event.Type == alive.EventResult {
		ts.task.Results = upsertResult(ts.task.Results, timestampResult(res, now))
		if len(res.AttemptResults) == 0 {
			ts.prependLogLocked(now, res)
		}
		ts.completeRunLocked(runID, res.Model, now)
		ts.publishLocked()
	}
	return true
}

func (ts *taskStore) finishRun(taskID, runID int64, model string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.isCurrentRunLocked(runID, model) {
		return
	}
	if ts.task.ID != taskID {
		if ts.completeRunLocked(runID, model, time.Now().Format(time.RFC3339)) {
			ts.publishLocked()
		}
		return
	}
	if ts.completeRunLocked(runID, model, time.Now().Format(time.RFC3339)) {
		ts.publishLocked()
	}
}

func (ts *taskStore) failRun(taskID, runID int64, model, message string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.isCurrentRunLocked(runID, model) {
		return
	}
	if ts.task.ID != taskID {
		if ts.completeRunLocked(runID, model, time.Now().Format(time.RFC3339)) {
			ts.publishLocked()
		}
		return
	}
	delete(ts.runs, model)
	ts.task.RunningModels = removeModelName(ts.task.RunningModels, model)
	ts.task.Error = message
	if len(ts.runs) > 0 {
		ts.publishLocked()
		return
	}
	ts.task.Running = false
	ts.task.RunningModels = nil
	ts.task.FinishedAt = time.Now().Format(time.RFC3339)
	ts.runs = nil
	ts.publishLocked()
}

func (ts *taskStore) isCurrentRunLocked(runID int64, model string) bool {
	run, ok := ts.runs[model]
	return ok && run.ID == runID
}

func (ts *taskStore) runIsCurrent(taskID, runID int64, model string) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.task.ID == taskID && ts.isCurrentRunLocked(runID, model)
}

func (ts *taskStore) reconcileRunningLocked() {
	if len(ts.runs) == 0 {
		if ts.task.Running {
			ts.task.Running = false
			ts.task.RunningModels = nil
		}
		return
	}
	ts.task.Running = true
	ts.task.FinishedAt = ""
	ts.task.Models = alive.AddModels(ts.task.Models, activeRunModels(ts.runs))
	ts.task.RunningModels = runningModelsInOrder(ts.task.Models, ts.runs)
}

func (ts *taskStore) completeRunLocked(runID int64, model, finishedAt string) bool {
	if !ts.isCurrentRunLocked(runID, model) {
		return false
	}
	delete(ts.runs, model)
	ts.task.RunningModels = removeModelName(ts.task.RunningModels, model)
	if len(ts.runs) > 0 {
		return true
	}
	ts.task.Running = false
	ts.task.RunningModels = nil
	ts.task.FinishedAt = finishedAt
	ts.runs = nil
	return true
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

func activeRunModels(runs map[string]activeModelRun) []string {
	models := make([]string, 0, len(runs))
	for model := range runs {
		models = append(models, model)
	}
	return models
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

func resultByModelName(results []alive.Result, model string) alive.Result {
	for _, res := range results {
		if res.Model == model {
			return res
		}
	}
	return alive.Result{Model: model}
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
