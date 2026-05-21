// internal/transfer/transfer.go

package transfer

import (
	"fmt"
	"strings"
	"time"

	"github.com/riteshmishra/llama-nest/internal/db"
	"github.com/riteshmishra/llama-nest/internal/ollama"
	"github.com/riteshmishra/llama-nest/internal/types"
)

type Service struct {
	Store  *db.Store
	Ollama *ollama.Client
}

type Result struct {
	TransferID      int64
	SourceModel     string
	TargetModel     string
	Acknowledgement string
}

func NewService(ollamaURL string, store *db.Store) *Service {
	return &Service{
		Store:  store,
		Ollama: ollama.New(ollamaURL),
	}
}

func (s *Service) Transfer(targetModel string) (Result, error) {
	targetModel = strings.TrimSpace(targetModel)
	if targetModel == "" {
		return Result{}, fmt.Errorf("target model is required")
	}

	exists, err := s.Ollama.ModelExists(targetModel)
	if err != nil {
		return Result{}, err
	}

	if !exists {
		if err := s.Ollama.Pull(targetModel); err != nil {
			return Result{}, err
		}
	}

	msgs, err := s.Store.RecentMessages(30)
	if err != nil {
		return Result{}, err
	}

	if len(msgs) == 0 {
		return Result{}, fmt.Errorf("no captured messages found; send traffic through the llama-nest proxy first")
	}

	sourceModel := latestModel(msgs)
	pack := buildPack(sourceModel, targetModel, msgs)

	ack, err := s.Ollama.Chat(targetModel, []map[string]string{
		{
			"role":    "system",
			"content": transferSystemPrompt(),
		},
		{
			"role":    "user",
			"content": pack.Brief,
		},
	})
	if err != nil {
		return Result{}, err
	}

	id, err := s.Store.CreateTransfer(sourceModel, targetModel, pack, ack)
	if err != nil {
		return Result{}, err
	}

	return Result{
		TransferID:      id,
		SourceModel:     sourceModel,
		TargetModel:     targetModel,
		Acknowledgement: ack,
	}, nil
}

func buildPack(sourceModel, targetModel string, msgs []types.Message) types.TransferPack {
	var b strings.Builder

	b.WriteString("# llama-nest transfer pack\n\n")
	b.WriteString("You are being transferred into an existing local AI session.\n")
	b.WriteString("Your job is to get caught up, preserve context, and be ready to continue.\n\n")
	b.WriteString("Source model: " + sourceModel + "\n")
	b.WriteString("Target model: " + targetModel + "\n\n")
	b.WriteString("Recent context:\n\n")

	for _, m := range msgs {
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}

		if len(content) > 1000 {
			content = content[:1000] + "..."
		}

		b.WriteString("- " + m.Role + ": " + content + "\n")
	}

	return types.TransferPack{
		SourceModel:    sourceModel,
		TargetModel:    targetModel,
		CreatedAt:      time.Now().UTC(),
		RecentMessages: msgs,
		Brief:          b.String(),
	}
}

func transferSystemPrompt() string {
	return `You are being transferred into an existing session.

Read the provided context pack. Do not continue the user conversation yet.

Acknowledge only with:
- current objective
- relevant constraints
- what context must be preserved
- next best action`
}

func latestModel(msgs []types.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if strings.TrimSpace(msgs[i].Model) != "" {
			return msgs[i].Model
		}
	}

	return "unknown"
}
