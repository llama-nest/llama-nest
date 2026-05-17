package types

import "time"

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
