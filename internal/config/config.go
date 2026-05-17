package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	DataDir     string `json:"data_dir"`
	DatabaseURL string `json:"database_url"`
	OllamaURL   string `json:"ollama_url"`
	ProxyAddr   string `json:"proxy_addr"`
	APIAddr     string `json:"api_addr"`
}

func Default() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	dataDir := filepath.Join(home, ".llama-nest")
	return Config{
		DataDir:     dataDir,
		DatabaseURL: filepath.Join(dataDir, "llama-nest.db"),
		OllamaURL:   env("LLAMA_NEST_OLLAMA_URL", "http://localhost:11434"),
		ProxyAddr:   env("LLAMA_NEST_PROXY_ADDR", ":11435"),
		APIAddr:     env("LLAMA_NEST_API_ADDR", ":8787"),
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".llama-nest", "config.json"), nil
}

func Load() (Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Default()
	}
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.OllamaURL == "" {
		c.OllamaURL = "http://localhost:11434"
	}
	if c.ProxyAddr == "" {
		c.ProxyAddr = ":11435"
	}
	if c.APIAddr == "" {
		c.APIAddr = ":8787"
	}
	return c, nil
}

func Save(c Config) error {
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return err
	}
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}
