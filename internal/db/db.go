package db

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/riteshmishra/llama-nest/internal/types"
)

type Store struct {
	Dir string
	mu  sync.Mutex
}

type dataFile struct {
	Sessions []types.Session `json:"sessions"`
	Messages []types.Message `json:"messages"`
}

func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return &Store{Dir: dir}, nil
}

func (s *Store) Migrate() error       { return os.MkdirAll(s.Dir, 0700) }
func (s *Store) sessionsPath() string { return filepath.Join(s.Dir, "sessions.jsonl") }
func (s *Store) messagesPath() string { return filepath.Join(s.Dir, "messages.jsonl") }

func (s *Store) readSessions() ([]types.Session, error) {
	f, err := os.Open(s.sessionsPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []types.Session
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 1024*1024*10)
	for sc.Scan() {
		var x types.Session
		if json.Unmarshal(sc.Bytes(), &x) == nil {
			out = append(out, x)
		}
	}
	return out, sc.Err()
}
func (s *Store) readMessages() ([]types.Message, error) {
	f, err := os.Open(s.messagesPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []types.Message
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 1024*1024*10)
	for sc.Scan() {
		var x types.Message
		if json.Unmarshal(sc.Bytes(), &x) == nil {
			out = append(out, x)
		}
	}
	return out, sc.Err()
}
func appendJSONL(path string, v any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func (s *Store) CreateSession(model, title string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessions, err := s.readSessions()
	if err != nil {
		return 0, err
	}
	id := int64(len(sessions) + 1)
	now := time.Now().UTC()
	if strings.TrimSpace(title) == "" {
		title = "Untitled session"
	}
	x := types.Session{ID: id, Model: model, Title: title, CreatedAt: now, UpdatedAt: now}
	return id, appendJSONL(s.sessionsPath(), x)
}
func (s *Store) AddMessage(sessionID int64, role, content, model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs, err := s.readMessages()
	if err != nil {
		return err
	}
	x := types.Message{ID: int64(len(msgs) + 1), SessionID: sessionID, Role: role, Content: content, Model: model, CreatedAt: time.Now().UTC()}
	return appendJSONL(s.messagesPath(), x)
}
func (s *Store) Capture(model string, msgs []types.Message, response string) (int64, error) {
	title := "Ollama session"
	for _, m := range msgs {
		if m.Role == "user" && strings.TrimSpace(m.Content) != "" {
			title = truncate(m.Content, 72)
			break
		}
	}
	sid, err := s.CreateSession(model, title)
	if err != nil {
		return 0, err
	}
	for _, m := range msgs {
		if err := s.AddMessage(sid, m.Role, m.Content, model); err != nil {
			return 0, err
		}
	}
	if strings.TrimSpace(response) != "" {
		if err := s.AddMessage(sid, "assistant", response, model); err != nil {
			return 0, err
		}
	}
	return sid, nil
}
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
func (s *Store) Status() (map[string]int64, error) {
	sessions, err := s.readSessions()
	if err != nil {
		return nil, err
	}

	msgs, err := s.readMessages()
	if err != nil {
		return nil, err
	}

	transfers, err := s.readTransfers()
	if err != nil {
		return nil, err
	}

	usage, err := s.readUsage()
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"sessions":  int64(len(sessions)),
		"messages":  int64(len(msgs)),
		"transfers": int64(len(transfers)),
		"usage":     int64(len(usage)),
	}, nil
}
func (s *Store) Sessions(limit int) ([]types.Session, error) {
	xs, err := s.readSessions()
	if err != nil {
		return nil, err
	}
	reverseSessions(xs)
	if limit > 0 && len(xs) > limit {
		xs = xs[:limit]
	}
	return xs, nil
}
func (s *Store) Messages(sessionID int64) ([]types.Message, error) {
	xs, err := s.readMessages()
	if err != nil {
		return nil, err
	}
	var out []types.Message
	for _, m := range xs {
		if m.SessionID == sessionID {
			out = append(out, m)
		}
	}
	return out, nil
}
func (s *Store) Search(q string, limit int) ([]types.Message, error) {
	xs, err := s.readMessages()
	if err != nil {
		return nil, err
	}
	q = strings.ToLower(strings.TrimSpace(q))
	var out []types.Message
	for i := len(xs) - 1; i >= 0; i-- {
		if strings.Contains(strings.ToLower(xs[i].Content), q) {
			out = append(out, xs[i])
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (s *Store) RecentMessages(limit int) ([]types.Message, error) {
	msgs, err := s.readMessages()
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(msgs) {
		limit = len(msgs)
	}

	start := len(msgs) - limit
	if start < 0 {
		start = 0
	}

	return msgs[start:], nil
}
func reverseSessions(xs []types.Session) {
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
}
func (s *Store) transfersPath() string {
	return filepath.Join(s.Dir, "transfers.jsonl")
}
func (s *Store) readTransfers() ([]types.Transfer, error) {
	f, err := os.Open(s.transfersPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []types.Transfer
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 1024*1024*10)

	for sc.Scan() {
		var x types.Transfer
		if json.Unmarshal(sc.Bytes(), &x) == nil {
			out = append(out, x)
		}
	}

	return out, sc.Err()
}
func (s *Store) CreateTransfer(sourceModel, targetModel string, pack types.TransferPack, acknowledgement string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transfers, err := s.readTransfers()
	if err != nil {
		return 0, err
	}

	x := types.Transfer{
		ID:              int64(len(transfers) + 1),
		SourceModel:     sourceModel,
		TargetModel:     targetModel,
		Pack:            pack,
		Acknowledgement: acknowledgement,
		CreatedAt:       time.Now().UTC(),
	}

	return x.ID, appendJSONL(s.transfersPath(), x)
}
func (s *Store) Transfers(limit int) ([]types.Transfer, error) {
	xs, err := s.readTransfers()
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}

	if limit > 0 && len(xs) > limit {
		xs = xs[:limit]
	}

	return xs, nil
}
func (s *Store) usagePath() string {
	return filepath.Join(s.Dir, "usage.jsonl")
}
func (s *Store) readUsage() ([]types.Usage, error) {
	f, err := os.Open(s.usagePath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []types.Usage
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 1024*1024*10)

	for sc.Scan() {
		var x types.Usage
		if json.Unmarshal(sc.Bytes(), &x) == nil {
			out = append(out, x)
		}
	}

	return out, sc.Err()
}
func (s *Store) CreateUsage(sessionID int64, model, endpoint string, promptTokens, completionTokens int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.readUsage()
	if err != nil {
		return 0, err
	}

	x := types.Usage{
		ID:               int64(len(rows) + 1),
		SessionID:        sessionID,
		Model:            model,
		Endpoint:         endpoint,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		CreatedAt:        time.Now().UTC(),
	}

	return x.ID, appendJSONL(s.usagePath(), x)
}
func (s *Store) Usage(limit int) ([]types.Usage, error) {
	xs, err := s.readUsage()
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}

	if limit > 0 && len(xs) > limit {
		xs = xs[:limit]
	}

	return xs, nil
}
func (s *Store) UsageSummary() (types.UsageSummary, error) {
	rows, err := s.readUsage()
	if err != nil {
		return types.UsageSummary{}, err
	}

	summary := types.UsageSummary{}
	byModel := map[string]*types.UsageByModel{}

	for _, row := range rows {
		summary.PromptTokens += row.PromptTokens
		summary.CompletionTokens += row.CompletionTokens
		summary.TotalTokens += row.TotalTokens

		key := row.Model
		if key == "" {
			key = "unknown"
		}

		if _, ok := byModel[key]; !ok {
			byModel[key] = &types.UsageByModel{Model: key}
		}

		byModel[key].PromptTokens += row.PromptTokens
		byModel[key].CompletionTokens += row.CompletionTokens
		byModel[key].TotalTokens += row.TotalTokens
	}

	for _, v := range byModel {
		summary.ByModel = append(summary.ByModel, *v)
	}

	return summary, nil
}
func (s *Store) Export() (types.NestExport, error) {
	sessions, err := s.readSessions()
	if err != nil {
		return types.NestExport{}, err
	}

	messages, err := s.readMessages()
	if err != nil {
		return types.NestExport{}, err
	}

	transfers, err := s.readTransfers()
	if err != nil {
		return types.NestExport{}, err
	}

	usage, err := s.readUsage()
	if err != nil {
		return types.NestExport{}, err
	}

	return types.NestExport{
		Version:    "0.1",
		ExportedAt: time.Now().UTC(),
		Sessions:   sessions,
		Messages:   messages,
		Transfers:  transfers,
		Usage:      usage,
	}, nil
}
func (s *Store) Wipe() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	paths := []string{
		s.sessionsPath(),
		s.messagesPath(),
		s.transfersPath(),
		s.usagePath(),
	}

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}
