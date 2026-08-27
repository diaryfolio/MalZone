package siem

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"sync"
)

type Sink struct {
	mu     sync.RWMutex
	events map[string]ECSEvent
	limit  int
}

func NewSink(limit int) http.Handler {
	if limit <= 0 {
		limit = 100
	}
	sink := &Sink{events: make(map[string]ECSEvent), limit: limit}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeSinkJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /events", sink.receive)
	mux.HandleFunc("GET /events", sink.list)
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, request)
	})
}

func (s *Sink) receive(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, 16*1024)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var event ECSEvent
	if err := decoder.Decode(&event); err != nil || event.Event.ID == "" || event.Timestamp == "" {
		writeSinkJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ECS event"})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeSinkJSON(w, http.StatusBadRequest, map[string]string{"error": "one ECS event is required"})
		return
	}
	s.mu.Lock()
	if _, exists := s.events[event.Event.ID]; !exists && len(s.events) >= s.limit {
		s.mu.Unlock()
		writeSinkJSON(w, http.StatusInsufficientStorage, map[string]string{"error": "POC sink capacity reached"})
		return
	}
	s.events[event.Event.ID] = event
	s.mu.Unlock()
	writeSinkJSON(w, http.StatusAccepted, map[string]string{"eventId": event.Event.ID})
}

func (s *Sink) list(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	items := make([]ECSEvent, 0, len(s.events))
	for _, event := range s.events {
		items = append(items, event)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].Event.ID < items[j].Event.ID })
	writeSinkJSON(w, http.StatusOK, map[string]any{"items": items})
}

func writeSinkJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
