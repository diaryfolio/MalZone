package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/diaryfolio/malzone/internal/kube"
	"github.com/diaryfolio/malzone/internal/model"
)

type Store interface {
	ListAnalyses(context.Context) ([]model.Analysis, error)
	GetAnalysis(context.Context, string) (model.Analysis, error)
	CreateAnalysis(context.Context, model.Analysis) (model.Analysis, error)
	RequestCancel(context.Context, string) (model.Analysis, error)
	AppendInteraction(context.Context, string, model.InteractionAction) (model.Analysis, error)
}

type Server struct {
	store     Store
	namespace string
	logger    *slog.Logger
}

func New(store Store, namespace string, logger *slog.Logger) http.Handler {
	server := &Server{store: store, namespace: namespace, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.health)
	mux.HandleFunc("GET /readyz", server.ready)
	mux.HandleFunc("GET /api/v1alpha1/analyses", server.list)
	mux.HandleFunc("POST /api/v1alpha1/analyses", server.create)
	mux.HandleFunc("GET /api/v1alpha1/analyses/{name}", server.get)
	mux.HandleFunc("POST /api/v1alpha1/analyses/{name}/cancel", server.cancel)
	mux.HandleFunc("POST /api/v1alpha1/analyses/{name}/actions", server.createAction)
	return requestLog(logger, securityHeaders(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()
	if _, err := s.store.ListAnalyses(ctx); err != nil {
		problem(w, http.StatusServiceUnavailable, "kubernetes_unavailable", "Kubernetes API is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) list(w http.ResponseWriter, request *http.Request) {
	items, err := s.store.ListAnalyses(request.Context())
	if err != nil {
		s.internalError(w, request, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"apiVersion": "malzone.poc/v1alpha1", "items": items})
}

func (s *Server) create(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, 4096)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var create model.CreateAnalysisRequest
	if err := decoder.Decode(&create); err != nil {
		problem(w, http.StatusBadRequest, "invalid_request", "Request body must be valid bounded JSON")
		return
	}
	if err := ensureEOF(decoder); err != nil {
		problem(w, http.StatusBadRequest, "invalid_request", "Request body must contain one JSON object")
		return
	}
	if create.Name == "" {
		name, err := generatedName()
		if err != nil {
			s.internalError(w, request, err)
			return
		}
		create.Name = name
	}
	if err := model.ValidateCreate(create); err != nil {
		problem(w, http.StatusUnprocessableEntity, "invalid_analysis", err.Error())
		return
	}
	created, err := s.store.CreateAnalysis(request.Context(), model.NewAnalysis(s.namespace, create))
	if err != nil {
		if kube.IsConflict(err) {
			problem(w, http.StatusConflict, "analysis_exists", "An analysis with this name already exists")
			return
		}
		s.internalError(w, request, err)
		return
	}
	w.Header().Set("Location", "/api/v1alpha1/analyses/"+created.Metadata.Name)
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) get(w http.ResponseWriter, request *http.Request) {
	name := request.PathValue("name")
	if err := model.ValidateName(name); err != nil {
		problem(w, http.StatusBadRequest, "invalid_name", "Analysis name is invalid")
		return
	}
	analysis, err := s.store.GetAnalysis(request.Context(), name)
	if err != nil {
		if kube.IsNotFound(err) {
			problem(w, http.StatusNotFound, "analysis_not_found", "Analysis was not found")
			return
		}
		s.internalError(w, request, err)
		return
	}
	writeJSON(w, http.StatusOK, analysis)
}

func (s *Server) cancel(w http.ResponseWriter, request *http.Request) {
	name := request.PathValue("name")
	if err := model.ValidateName(name); err != nil {
		problem(w, http.StatusBadRequest, "invalid_name", "Analysis name is invalid")
		return
	}
	analysis, err := s.store.RequestCancel(request.Context(), name)
	if err != nil {
		if kube.IsNotFound(err) {
			problem(w, http.StatusNotFound, "analysis_not_found", "Analysis was not found")
			return
		}
		s.internalError(w, request, err)
		return
	}
	writeJSON(w, http.StatusAccepted, analysis)
}

func (s *Server) createAction(w http.ResponseWriter, request *http.Request) {
	name := request.PathValue("name")
	if err := model.ValidateName(name); err != nil {
		problem(w, http.StatusBadRequest, "invalid_name", "Analysis name is invalid")
		return
	}
	request.Body = http.MaxBytesReader(w, request.Body, 2048)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var create model.CreateInteractionRequest
	if err := decoder.Decode(&create); err != nil || ensureEOF(decoder) != nil {
		problem(w, http.StatusBadRequest, "invalid_request", "Request body must contain one bounded JSON object")
		return
	}
	if err := model.ValidateInteraction(create); err != nil {
		problem(w, http.StatusUnprocessableEntity, "invalid_action", err.Error())
		return
	}
	analysis, err := s.store.GetAnalysis(request.Context(), name)
	if err != nil {
		if kube.IsNotFound(err) {
			problem(w, http.StatusNotFound, "analysis_not_found", "Analysis was not found")
			return
		}
		s.internalError(w, request, err)
		return
	}
	if analysis.Status.Phase != "Running" || analysis.Spec.CancelRequested {
		problem(w, http.StatusConflict, "analysis_not_interactive", "POC observations require an active Running analysis")
		return
	}
	if len(analysis.Spec.Interactions) >= 20 {
		problem(w, http.StatusTooManyRequests, "action_budget_exhausted", "POC observation budget is 20 actions")
		return
	}
	id, err := generatedID("action-")
	if err != nil {
		s.internalError(w, request, err)
		return
	}
	action := model.NewInteraction(id, create, time.Now())
	updated, err := s.store.AppendInteraction(request.Context(), name, action)
	if err != nil {
		if kube.IsConflict(err) {
			problem(w, http.StatusConflict, "action_conflict", "Analysis changed; retry against current state")
			return
		}
		s.internalError(w, request, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"action": action, "analysis": updated})
}

func (s *Server) internalError(w http.ResponseWriter, request *http.Request, err error) {
	s.logger.Error("request failed", "method", request.Method, "path", request.URL.Path, "error", err)
	problem(w, http.StatusInternalServerError, "internal_error", "The request could not be completed")
}

func generatedName() (string, error) {
	return generatedID("analysis-")
}

func generatedID(prefix string) (string, error) {
	value := make([]byte, 6)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate identifier: %w", err)
	}
	return prefix + hex.EncodeToString(value), nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("additional JSON value")
	}
	return err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func problem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "urn:malzone:problem:" + code,
		"title":  http.StatusText(status),
		"status": status,
		"code":   code,
		"detail": detail,
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, request)
	})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, request)
		logger.Info("http request", "method", request.Method, "route", routeTemplate(request),
			"duration_ms", time.Since(started).Milliseconds())
	})
}

func routeTemplate(request *http.Request) string {
	path := request.URL.Path
	if name := request.PathValue("name"); name != "" {
		path = strings.Replace(path, name, "{name}", 1)
	}
	return path
}
