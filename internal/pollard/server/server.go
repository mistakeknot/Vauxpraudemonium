package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mistakeknot/autarch/internal/pollard/api"
	"github.com/mistakeknot/autarch/internal/pollard/hunters"
	"github.com/mistakeknot/autarch/internal/pollard/insights"
	"github.com/mistakeknot/autarch/pkg/httpapi"
	"github.com/mistakeknot/autarch/pkg/netguard"
)

const (
	defaultCacheMax = 512
	defaultJobsMax  = 20000
	defaultJobsTTL  = 24 * time.Hour
)

type Server struct {
	root    string
	scanner *api.Scanner
	cache   *ScanCache
	jobs    *JobStore
	mux     *http.ServeMux
	srv     *http.Server
}

func New(root string) (*Server, error) {
	scanner, err := api.NewScanner(root)
	if err != nil {
		return nil, err
	}
	s := &Server{
		root:    root,
		scanner: scanner,
		cache:   NewScanCache(defaultCacheMax),
		jobs:    NewJobStore(defaultJobsTTL, defaultJobsMax),
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) Close() error {
	if s.scanner != nil {
		return s.scanner.Close()
	}
	return nil
}

// Scanner exposes the underlying Pollard scanner for integrations.
func (s *Server) Scanner() *api.Scanner {
	return s.scanner
}

func (s *Server) ListenAndServe(addr string) error {
	if err := netguard.EnsureLocalOnly(addr); err != nil {
		return err
	}
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	return s.srv.ListenAndServe()
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/api/scan/targeted", s.handleTargetedScan)
	s.mux.HandleFunc("/api/research", s.handleResearch)
	s.mux.HandleFunc("/api/insights", s.handleInsights)
	s.mux.HandleFunc("/api/hunters", s.handleHunters)
	s.mux.HandleFunc("/api/jobs/", s.handleJobs)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	httpapi.WriteOK(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

type scanRequest struct {
	Hunters    []string               `json:"hunters"`
	Queries    []string               `json:"queries"`
	Targets    []api.CompetitorTarget `json:"targets"`
	MaxResults int                    `json:"max_results"`
	Mode       string                 `json:"mode,omitempty"`
}

type researchRequest struct {
	Vision       string   `json:"vision"`
	Problem      string   `json:"problem"`
	Requirements []string `json:"requirements"`
}

type targetedScanRequest struct {
	SpecID  string   `json:"spec_id"`
	Hunters []string `json:"hunters"`
	Mode    string   `json:"mode"`
	Query   string   `json:"query"`
}

type jobSummary struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type jobStatus struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Status     JobStatus  `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      string     `json:"error,omitempty"`
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, httpapi.ErrInvalidRequest, "invalid JSON body", nil, false)
		return
	}
	opts := api.ScanOptions{
		Hunters:    req.Hunters,
		Queries:    req.Queries,
		Targets:    req.Targets,
		MaxResults: req.MaxResults,
	}
	cacheTTL := ttlForMode(req.Mode)
	key := hashKey(struct {
		Endpoint string
		Opts     api.ScanOptions
		Mode     string
	}{Endpoint: "scan", Opts: opts, Mode: req.Mode})

	job := s.jobs.Create("scan")
	_ = s.jobs.Start(job.ID, func(ctx context.Context) (any, error) {
		val, err := s.cache.GetOrCompute(key, cacheTTL, func() (any, error) {
			result, err := s.scanner.Scan(ctx, opts)
			if err != nil {
				return nil, err
			}
			return toScanResult(result), nil
		})
		return val, err
	})

	httpapi.WriteOK(w, http.StatusAccepted, toJobSummary(job), nil)
}

func (s *Server) handleTargetedScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req targetedScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, httpapi.ErrInvalidRequest, "invalid JSON body", nil, false)
		return
	}
	opts := api.TargetedScanOpts{
		SpecID:  req.SpecID,
		Hunters: req.Hunters,
		Mode:    api.ScanMode(req.Mode),
		Query:   req.Query,
	}
	cacheTTL := ttlForMode(req.Mode)
	key := hashKey(struct {
		Endpoint string
		Opts     api.TargetedScanOpts
	}{Endpoint: "targeted", Opts: opts})

	job := s.jobs.Create("scan_targeted")
	_ = s.jobs.Start(job.ID, func(ctx context.Context) (any, error) {
		val, err := s.cache.GetOrCompute(key, cacheTTL, func() (any, error) {
			result, err := s.scanner.RunTargetedScan(ctx, opts)
			if err != nil {
				return nil, err
			}
			return toTargetedResult(result), nil
		})
		return val, err
	})

	httpapi.WriteOK(w, http.StatusAccepted, toJobSummary(job), nil)
}

func (s *Server) handleResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req researchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, httpapi.ErrInvalidRequest, "invalid JSON body", nil, false)
		return
	}
	job := s.jobs.Create("research")
	_ = s.jobs.Start(job.ID, func(ctx context.Context) (any, error) {
		result, err := s.scanner.ResearchForPRD(ctx, req.Vision, req.Problem, req.Requirements)
		if err != nil {
			return nil, err
		}
		return toScanResult(result), nil
	})

	httpapi.WriteOK(w, http.StatusAccepted, toJobSummary(job), nil)
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	list, err := insights.LoadAll(s.root)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, httpapi.ErrInternal, "failed to load insights", nil, false)
		return
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CollectedAt.After(list[j].CollectedAt)
	})
	cursor, limit := parsePagination(r, 50)
	paged, next := paginate(list, cursor, limit)
	meta := &httpapi.Meta{Cursor: next, Limit: limit}
	httpapi.WriteOK(w, http.StatusOK, paged, meta)
}

func (s *Server) handleHunters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	var names []string
	for _, h := range s.scanner.AvailableHunters() {
		names = append(names, h)
	}
	if len(names) == 0 {
		// Fallback to default registry if scanner hasn't loaded yet.
		reg := hunters.DefaultRegistry()
		for _, h := range reg.All() {
			names = append(names, h.Name())
		}
	}

	sort.Strings(names)
	httpapi.WriteOK(w, http.StatusOK, map[string][]string{"hunters": names}, nil)
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	path = strings.Trim(path, "/")
	if path == "" {
		httpapi.WriteError(w, http.StatusNotFound, httpapi.ErrNotFound, "job not found", nil, false)
		return
	}
	parts := strings.Split(path, "/")
	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		job, ok := s.jobs.Get(id)
		if !ok {
			httpapi.WriteError(w, http.StatusNotFound, httpapi.ErrNotFound, "job not found", nil, false)
			return
		}
		httpapi.WriteOK(w, http.StatusOK, toJobStatus(job), nil)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "result":
			if r.Method != http.MethodGet {
				methodNotAllowed(w)
				return
			}
			s.handleJobResult(w, id)
			return
		case "cancel":
			if r.Method != http.MethodPost {
				methodNotAllowed(w)
				return
			}
			job, err := s.jobs.Cancel(id)
			if err != nil {
				httpapi.WriteError(w, http.StatusConflict, httpapi.ErrConflict, err.Error(), nil, false)
				return
			}
			httpapi.WriteOK(w, http.StatusOK, toJobStatus(job), nil)
			return
		}
	}
	httpapi.WriteError(w, http.StatusNotFound, httpapi.ErrNotFound, "job not found", nil, false)
}

func (s *Server) handleJobResult(w http.ResponseWriter, id string) {
	job, ok := s.jobs.Get(id)
	if !ok {
		httpapi.WriteError(w, http.StatusNotFound, httpapi.ErrNotFound, "job not found", nil, false)
		return
	}
	switch job.Status {
	case JobSucceeded:
		httpapi.WriteOK(w, http.StatusOK, job.Result, nil)
	case JobFailed, JobCanceled, JobExpired:
		httpapi.WriteError(w, http.StatusConflict, httpapi.ErrJobFailed, job.Error, nil, false)
	default:
		httpapi.WriteError(w, http.StatusConflict, httpapi.ErrJobPending, "job not complete", nil, false)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusMethodNotAllowed, httpapi.ErrInvalidRequest, "method not allowed", nil, false)
}

func parsePagination(r *http.Request, defaultLimit int) (int, int) {
	cursor := 0
	if v := r.URL.Query().Get("cursor"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			cursor = parsed
		}
	}
	limit := defaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return cursor, limit
}

func paginate[T any](items []T, cursor int, limit int) ([]T, string) {
	if cursor >= len(items) {
		return []T{}, ""
	}
	end := cursor + limit
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = strconv.Itoa(end)
	}
	return items[cursor:end], next
}

func ttlForMode(mode string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "quick":
		return 5 * time.Minute
	case "deep":
		return 30 * time.Minute
	case "balanced":
		return 15 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func toJobSummary(job *Job) jobSummary {
	return jobSummary{
		ID:        job.ID,
		Type:      job.Type,
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
}

func toJobStatus(job *Job) jobStatus {
	return jobStatus{
		ID:         job.ID,
		Type:       job.Type,
		Status:     job.Status,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
		Error:      job.Error,
	}
}

// Result DTOs

type huntResultDTO struct {
	HunterName       string    `json:"hunter_name"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	SourcesCollected int       `json:"sources_collected"`
	InsightsCreated  int       `json:"insights_created"`
	OutputFiles      []string  `json:"output_files"`
	Errors           []string  `json:"errors,omitempty"`
}

type scanResultDTO struct {
	HunterResults map[string]huntResultDTO `json:"hunter_results"`
	TotalSources  int                      `json:"total_sources"`
	TotalInsights int                      `json:"total_insights"`
	OutputFiles   []string                 `json:"output_files"`
	Errors        []string                 `json:"errors,omitempty"`
}

type targetedResultDTO struct {
	SpecID        string   `json:"spec_id"`
	Mode          string   `json:"mode"`
	Hunters       []string `json:"hunters"`
	TotalSources  int      `json:"total_sources"`
	TotalInsights int      `json:"total_insights"`
	OutputFiles   []string `json:"output_files"`
	Errors        []string `json:"errors,omitempty"`
}

func toScanResult(result *api.ScanResult) scanResultDTO {
	out := scanResultDTO{
		HunterResults: make(map[string]huntResultDTO),
		TotalSources:  result.TotalSources,
		TotalInsights: result.TotalInsights,
		OutputFiles:   append([]string{}, result.OutputFiles...),
		Errors:        errorsToStrings(result.Errors),
	}
	for name, hr := range result.HunterResults {
		out.HunterResults[name] = huntResultDTO{
			HunterName:       hr.HunterName,
			StartedAt:        hr.StartedAt,
			CompletedAt:      hr.CompletedAt,
			SourcesCollected: hr.SourcesCollected,
			InsightsCreated:  hr.InsightsCreated,
			OutputFiles:      append([]string{}, hr.OutputFiles...),
			Errors:           errorsToStrings(hr.Errors),
		}
	}
	return out
}

func toTargetedResult(result *api.TargetedScanResult) targetedResultDTO {
	return targetedResultDTO{
		SpecID:        result.SpecID,
		Mode:          string(result.Mode),
		Hunters:       append([]string{}, result.Hunters...),
		TotalSources:  result.TotalSources,
		TotalInsights: result.TotalInsights,
		OutputFiles:   append([]string{}, result.OutputFiles...),
		Errors:        errorsToStrings(result.Errors),
	}
}

func errorsToStrings(errs []error) []string {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		out = append(out, err.Error())
	}
	return out
}

// Ensure we don't accidentally return raw errors in JSON.
