package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default() returned error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedDataDir := filepath.Join(home, ".llama-nest")
	if cfg.DataDir != expectedDataDir {
		t.Errorf("DataDir = %s, want %s", cfg.DataDir, expectedDataDir)
	}

	expectedDBURL := filepath.Join(expectedDataDir, "llama-nest.db")
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("DatabaseURL = %s, want %s", cfg.DatabaseURL, expectedDBURL)
	}

	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL = %s, want http://localhost:11434", cfg.OllamaURL)
	}

	if cfg.ProxyAddr != ":11435" {
		t.Errorf("ProxyAddr = %s, want :11435", cfg.ProxyAddr)
	}

	if cfg.APIAddr != ":8787" {
		t.Errorf("APIAddr = %s, want :8787", cfg.APIAddr)
	}
}

func TestDefaultWithEnvVars(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("LLAMA_NEST_OLLAMA_URL")
		os.Unsetenv("LLAMA_NEST_PROXY_ADDR")
		os.Unsetenv("LLAMA_NEST_API_ADDR")
	})

	os.Setenv("LLAMA_NEST_OLLAMA_URL", "http://custom-ollama:11434")
	os.Setenv("LLAMA_NEST_PROXY_ADDR", ":9999")
	os.Setenv("LLAMA_NEST_API_ADDR", ":9090")

	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default() returned error: %v", err)
	}

	if cfg.OllamaURL != "http://custom-ollama:11434" {
		t.Errorf("OllamaURL = %s, want http://custom-ollama:11434", cfg.OllamaURL)
	}

	if cfg.ProxyAddr != ":9999" {
		t.Errorf("ProxyAddr = %s, want :9999", cfg.ProxyAddr)
	}

	if cfg.APIAddr != ":9090" {
		t.Errorf("APIAddr = %s, want :9090", cfg.APIAddr)
	}
}

func TestEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		want     string
		setEnv   bool
	}{
		{
			name:     "returns env var when set",
			key:      "TEST_KEY",
			fallback: "default",
			envValue: "custom",
			want:     "custom",
			setEnv:   true,
		},
		{
			name:     "returns fallback when env var not set",
			key:      "NONEXISTENT_KEY",
			fallback: "default",
			want:     "default",
			setEnv:   false,
		},
		{
			name:     "returns fallback when env var is empty",
			key:      "EMPTY_KEY",
			fallback: "default",
			envValue: "",
			want:     "default",
			setEnv:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Unsetenv(tt.key)
			})

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			got := env(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("env(%s, %s) = %s, want %s", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() returned error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(home, ".llama-nest", "config.json")
	if path != expectedPath {
		t.Errorf("ConfigPath() = %s, want %s", path, expectedPath)
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	cfg := Config{
		DataDir:     filepath.Join(tmpDir, ".llama-nest"),
		DatabaseURL: filepath.Join(tmpDir, ".llama-nest", "llama-nest.db"),
		OllamaURL:   "http://test-ollama:11434",
		ProxyAddr:   ":9999",
		APIAddr:     ":9090",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify the file was created
	configFile := filepath.Join(tmpDir, ".llama-nest", "config.json")
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load the config back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.OllamaURL != cfg.OllamaURL {
		t.Errorf("OllamaURL = %s, want %s", loaded.OllamaURL, cfg.OllamaURL)
	}

	if loaded.ProxyAddr != cfg.ProxyAddr {
		t.Errorf("ProxyAddr = %s, want %s", loaded.ProxyAddr, cfg.ProxyAddr)
	}

	if loaded.APIAddr != cfg.APIAddr {
		t.Errorf("APIAddr = %s, want %s", loaded.APIAddr, cfg.APIAddr)
	}
}

func TestLoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty for default config")
	}

	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL = %s, want http://localhost:11434", cfg.OllamaURL)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".llama-nest")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, []byte("invalid json {"), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestLoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	// Save a config with empty address fields
	configDir := filepath.Join(tmpDir, ".llama-nest")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	partialCfg := Config{DataDir: configDir}
	b, err := json.MarshalIndent(partialCfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, b, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load should fill in defaults for empty fields
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL = %s, want http://localhost:11434", cfg.OllamaURL)
	}

	if cfg.ProxyAddr != ":11435" {
		t.Errorf("ProxyAddr = %s, want :11435", cfg.ProxyAddr)
	}

	if cfg.APIAddr != ":8787" {
		t.Errorf("APIAddr = %s, want :8787", cfg.APIAddr)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	cfg := Config{
		DataDir:     filepath.Join(tmpDir, ".llama-nest"),
		DatabaseURL: filepath.Join(tmpDir, ".llama-nest", "llama-nest.db"),
		OllamaURL:   "http://localhost:11434",
		ProxyAddr:   ":11435",
		APIAddr:     ":8787",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(cfg.DataDir); err != nil {
		t.Fatalf("DataDir not created: %v", err)
	}
}

func TestSaveFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})

	os.Setenv("HOME", tmpDir)

	cfg := Config{
		DataDir:     filepath.Join(tmpDir, ".llama-nest"),
		DatabaseURL: filepath.Join(tmpDir, ".llama-nest", "llama-nest.db"),
		OllamaURL:   "http://localhost:11434",
		ProxyAddr:   ":11435",
		APIAddr:     ":8787",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	configFile := filepath.Join(tmpDir, ".llama-nest", "config.json")
	fileInfo, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check file permissions are 0600 (readable/writable by owner only)
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("File permissions = %o, want 0600", fileInfo.Mode().Perm())
	}
}

func TestConfigJSONMarshaling(t *testing.T) {
	cfg := Config{
		DataDir:     "/home/user/.llama-nest",
		DatabaseURL: "/home/user/.llama-nest/llama-nest.db",
		OllamaURL:   "http://localhost:11434",
		ProxyAddr:   ":11435",
		APIAddr:     ":8787",
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() returned error: %v", err)
	}

	var unmarshaled Config
	if err := json.Unmarshal(b, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() returned error: %v", err)
	}

	if unmarshaled != cfg {
		t.Errorf("Unmarshaled config = %+v, want %+v", unmarshaled, cfg)
	}
}
