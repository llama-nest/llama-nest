package types

import "time"

type TransferPack struct {
	SourceModel    string    `json:"source_model"`
	TargetModel    string    `json:"target_model"`
	CreatedAt      time.Time `json:"created_at"`
	RecentMessages []Message `json:"recent_messages"`
	Brief          string    `json:"brief"`
}

type Transfer struct {
	ID              int64        `json:"id"`
	SourceModel     string       `json:"source_model"`
	TargetModel     string       `json:"target_model"`
	Pack            TransferPack `json:"pack"`
	Acknowledgement string       `json:"acknowledgement"`
	CreatedAt       time.Time    `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	Model     string    `json:"model"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

type Capture struct {
	Model    string
	Messages []Message
	Response string
	Path     string
}

type Usage struct {
	ID               int64     `json:"id"`
	SessionID        int64     `json:"session_id"`
	Model            string    `json:"model"`
	Endpoint         string    `json:"endpoint"`
	PromptTokens     int64     `json:"prompt_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	TotalTokens      int64     `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

type UsageSummary struct {
	PromptTokens     int64          `json:"prompt_tokens"`
	CompletionTokens int64          `json:"completion_tokens"`
	TotalTokens      int64          `json:"total_tokens"`
	ByModel          []UsageByModel `json:"by_model"`
}

type UsageByModel struct {
	Model            string `json:"model"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

type NestExport struct {
	Version    string     `json:"version"`
	ExportedAt time.Time  `json:"exported_at"`
	Sessions   []Session  `json:"sessions"`
	Messages   []Message  `json:"messages"`
	Transfers  []Transfer `json:"transfers"`
	Usage      []Usage    `json:"usage"`
}
