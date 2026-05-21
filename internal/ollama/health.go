package ollama

import (
	"fmt"
	"net/http"
)

func (c *Client) Healthy() error {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not reachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ollama unhealthy: %s", resp.Status)
	}

	return nil
}
