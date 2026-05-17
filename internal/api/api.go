package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/riteshmishra/llama-nest/internal/db"
)

type Server struct{ Store *db.Store }

func New(store *db.Store) *Server { return &Server{Store: store} }

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.status)
	mux.HandleFunc("/api/sessions", s.sessions)
	mux.HandleFunc("/api/messages", s.messages)
	mux.HandleFunc("/api/search", s.search)
	mux.HandleFunc("/api/catch-up", s.catchUp)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("llama-nest api is running\n"))
	})
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func write(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, err error) { http.Error(w, err.Error(), 500) }

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	x, err := s.Store.Status()
	if err != nil {
		fail(w, err)
		return
	}
	write(w, x)
}
func (s *Server) sessions(w http.ResponseWriter, r *http.Request) {
	x, err := s.Store.Sessions(50)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, x)
}
func (s *Server) messages(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("session_id"), 10, 64)
	x, err := s.Store.Messages(id)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, x)
}
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	x, err := s.Store.Search(r.URL.Query().Get("q"), 25)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, x)
}
func (s *Server) catchUp(w http.ResponseWriter, r *http.Request) {
	msgs, err := s.Store.RecentMessages(30)
	if err != nil {
		fail(w, err)
		return
	}
	var b strings.Builder
	b.WriteString("# llama-nest catch-up brief\n\nRecent local context:\n\n")
	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		c := strings.TrimSpace(m.Content)
		if c == "" {
			continue
		}
		if len(c) > 500 {
			c = c[:500] + "…"
		}
		b.WriteString("- " + m.Role + ": " + c + "\n")
	}
	write(w, map[string]string{"brief": b.String()})
}
