package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/riteshmishra/llama-nest/internal/db"
	"github.com/riteshmishra/llama-nest/internal/types"
)

type Server struct {
	Target *url.URL
	Store  *db.Store
}

func New(target string, store *db.Store) (*Server, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return &Server{Target: u, Store: store}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	capture := parseCapture(r.URL.Path, body)

	target := *s.Target
	target.Path = singleJoiningSlash(s.Target.Path, r.URL.Path)
	target.RawQuery = r.URL.RawQuery
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "ollama unavailable: "+err.Error(), 502)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)

	if capture.Model != "" || len(capture.Messages) > 0 {
		answer := parseResponseText(respBody)
		if _, err := s.Store.Capture(capture.Model, capture.Messages, answer); err != nil {
			log.Printf("capture failed: %v", err)
		}
	}
}

type chatReq struct {
	Model    string                           `json:"model"`
	Messages []struct{ Role, Content string } `json:"messages"`
	Prompt   string                           `json:"prompt"`
}

func parseCapture(path string, body []byte) types.Capture {
	if !(strings.Contains(path, "/api/chat") || strings.Contains(path, "/api/generate") || strings.Contains(path, "/v1/chat/completions")) {
		return types.Capture{}
	}
	var req chatReq
	if err := json.Unmarshal(body, &req); err != nil {
		return types.Capture{}
	}
	var msgs []types.Message
	for _, m := range req.Messages {
		if strings.TrimSpace(m.Content) != "" {
			msgs = append(msgs, types.Message{Role: m.Role, Content: m.Content, Model: req.Model})
		}
	}
	if req.Prompt != "" {
		msgs = append(msgs, types.Message{Role: "user", Content: req.Prompt, Model: req.Model})
	}
	return types.Capture{Model: req.Model, Messages: msgs, Path: path}
}

func parseResponseText(body []byte) string {
	// Non-streaming Ollama chat: {"message":{"role":"assistant","content":"..."}}
	var a struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Response string `json:"response"`
		Choices  []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &a); err == nil {
		if a.Message.Content != "" {
			return a.Message.Content
		}
		if a.Response != "" {
			return a.Response
		}
		if len(a.Choices) > 0 {
			return a.Choices[0].Message.Content
		}
	}
	// Streaming NDJSON fallback: collect content fields best-effort.
	var b strings.Builder
	for _, line := range bytes.Split(body, []byte("\n")) {
		var x struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Response string `json:"response"`
		}
		if json.Unmarshal(line, &x) == nil {
			b.WriteString(x.Message.Content)
			b.WriteString(x.Response)
		}
	}
	return b.String()
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	default:
		return a + b
	}
}
