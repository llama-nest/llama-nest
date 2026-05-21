package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 10 * time.Minute},
	}
}

type tagResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (c *Client) ModelExists(model string) (bool, error) {
	resp, err := c.HTTP.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return false, fmt.Errorf("ollama tags failed: %s", resp.Status)
	}

	var out tagResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}

	for _, m := range out.Models {
		if m.Name == model {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) Pull(model string) error {
	body, _ := json.Marshal(map[string]any{
		"name":   model,
		"stream": false,
	})

	resp, err := c.HTTP.Post(c.BaseURL+"/api/pull", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ollama pull failed: %s", resp.Status)
	}

	return nil
}

func (c *Client) Chat(model string, messages []map[string]string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   false,
	})

	resp, err := c.HTTP.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama chat failed: %s", resp.Status)
	}

	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Message.Content, nil
}
