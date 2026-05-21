package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/riteshmishra/llama-nest/internal/api"
	"github.com/riteshmishra/llama-nest/internal/config"
	"github.com/riteshmishra/llama-nest/internal/db"
	"github.com/riteshmishra/llama-nest/internal/ollama"
	"github.com/riteshmishra/llama-nest/internal/proxy"
	"github.com/riteshmishra/llama-nest/internal/transfer"
)

func main() {
	cmd := "help"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var err error
	switch cmd {
	case "init":
		err = runInit()
	case "start":
		err = runStart()
	case "run":
		err = runChat(os.Args[2:])
	case "stop":
		err = runStop()
	case "status":
		err = runStatus()
	case "search":
		err = runSearch(os.Args[2:])
	case "transfer":
		err = runTransfer(os.Args[2:])
	case "usage":
		err = runUsage()
	case "export":
		err = runExport(os.Args[2:])
	case "wipe":
		err = runWipe(os.Args[2:])
	case "catch-up":
		err = runCatchUp()
	case "doctor":
		err = runDoctor()
	default:
		usage()
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`llama-nest - local-first memory sidecar for Ollama

Commands:
  init        one-time setup for local config and storage
	start       start the llama-nest sidecar API and Ollama proxy
	run MODEL   chat with a model through llama-nest capture
	stop        stop the running llama-nest sidecar
  search Q    search captured context
  transfer MODEL   transfer recent session context to another Ollama model
  usage           show local token usage
  export          export local context to a .nest file
  wipe        delete captured local context
  catch-up    print recent context brief
  doctor      check Ollama connectivity`)
}
func runChat(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("run requires a model")
	}

	c, _, err := loadStore()
	if err != nil {
		return err
	}

	model := strings.TrimSpace(args[0])
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("llama-nest chat")
	fmt.Println("model:", model)
	fmt.Println("type /exit or /quit to leave")
	fmt.Println()

	var messages []map[string]string

	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line == "/exit" || line == "/quit" {
			return nil
		}

		messages = append(messages, map[string]string{
			"role":    "user",
			"content": line,
		})

		reqBody, err := json.Marshal(map[string]any{
			"model":    model,
			"messages": messages,
			"stream":   false,
		})
		if err != nil {
			return err
		}

		proxyURL := c.ProxyAddr
		if strings.HasPrefix(proxyURL, ":") {
			proxyURL = "localhost" + proxyURL
		}

		resp, err := http.Post(
			"http://"+proxyURL+"/api/chat",
			"application/json",
			bytes.NewReader(reqBody),
		)
		if err != nil {
			return err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode >= 300 {
			return fmt.Errorf("proxy request failed: %s\n%s", resp.Status, string(respBody))
		}

		answer := parseChatAnswer(respBody)

		fmt.Println(answer)
		fmt.Println()

		messages = append(messages, map[string]string{
			"role":    "assistant",
			"content": answer,
		})
	}
}
func runUsage() error {
	_, s, err := loadStore()
	if err != nil {
		return err
	}

	x, err := s.UsageSummary()
	if err != nil {
		return err
	}

	fmt.Println("prompt tokens:", x.PromptTokens)
	fmt.Println("completion tokens:", x.CompletionTokens)
	fmt.Println("total tokens:", x.TotalTokens)

	if len(x.ByModel) > 0 {
		fmt.Println()
		fmt.Println("by model:")
		for _, m := range x.ByModel {
			fmt.Printf("- %s: %d total (%d prompt, %d completion)\n",
				m.Model,
				m.TotalTokens,
				m.PromptTokens,
				m.CompletionTokens,
			)
		}
	}

	return nil
}
func runExport(args []string) error {
	_, s, err := loadStore()
	if err != nil {
		return err
	}

	out := "llama-nest-context.nest"

	for i := 0; i < len(args); i++ {
		if args[i] == "--out" && i+1 < len(args) {
			out = args[i+1]
			i++
		}
	}

	x, err := s.Export()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, b, 0600); err != nil {
		return err
	}

	fmt.Println("exported local context to", out)
	return nil
}
func runInit() error {
	c, err := config.Default()
	if err != nil {
		return err
	}
	if err := config.Save(c); err != nil {
		return err
	}
	s, err := db.Open(c.DatabaseURL)
	if err != nil {
		return err
	}
	if err := s.Migrate(); err != nil {
		return err
	}
	fmt.Println("initialized llama-nest at", c.DataDir)
	return nil
}
func loadStore() (config.Config, *db.Store, error) {
	c, err := config.Load()
	if err != nil {
		return c, nil, err
	}
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return c, nil, err
	}
	s, err := db.Open(c.DatabaseURL)
	if err != nil {
		return c, nil, err
	}
	if err := s.Migrate(); err != nil {
		return c, nil, err
	}
	return c, s, nil
}
func runStart() error {
	c, s, err := loadStore()
	if err != nil {
		return err
	}
	pidPath := filepath.Join(c.DataDir, "llama-nest.pid")
	if _, err := os.Stat(pidPath); err == nil {
		return fmt.Errorf("llama-nest already running")
	}
	_ = os.WriteFile(
		pidPath,
		[]byte(strconv.Itoa(os.Getpid())),
		0600,
	)
	defer os.Remove(pidPath)
	ollamaClient := ollama.New(c.OllamaURL)
	if err := ollamaClient.Healthy(); err != nil {
		fmt.Println()
		fmt.Println("WARNING: Ollama does not appear to be running.")
		fmt.Println()
		fmt.Println("Start Ollama first:")
		fmt.Println("  ollama serve")
		fmt.Println()
		fmt.Println("llama-nest will still start, but model traffic will fail until Ollama is available.")
		fmt.Println()
	}
	apiSrv := api.New(s)
	px, err := proxy.New(c.OllamaURL, s)
	if err != nil {
		return err
	}
	go func() {
		log.Println("api listening on", c.APIAddr)
		if err := http.ListenAndServe(c.APIAddr, apiSrv.Routes()); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("proxy listening on", c.ProxyAddr, "->", c.OllamaURL)
	return http.ListenAndServe(c.ProxyAddr, px)
}
func runStop() error {
	c, err := config.Load()
	if err != nil {
		return err
	}

	pidPath := filepath.Join(c.DataDir, "llama-nest.pid")

	b, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("llama-nest is not running")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return err
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := p.Kill(); err != nil {
		return err
	}

	_ = os.Remove(pidPath)

	fmt.Println("Stopped llama-nest.")

	return nil
}
func runStatus() error {
	c, s, err := loadStore()
	if err != nil {
		return err
	}
	st, err := s.Status()
	if err != nil {
		return err
	}
	fmt.Println("data:", c.DataDir)
	fmt.Println("ollama:", c.OllamaURL)
	fmt.Println("api:", c.APIAddr)
	fmt.Println("proxy:", c.ProxyAddr)
	fmt.Println("sessions:", st["sessions"])
	fmt.Println("messages:", st["messages"])
	fmt.Println("transfers:", st["transfers"])
	return nil
}
func runSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search requires a query")
	}
	_, s, err := loadStore()
	if err != nil {
		return err
	}
	rows, err := s.Search(strings.Join(args, " "), 10)
	if err != nil {
		return err
	}
	for _, m := range rows {
		fmt.Printf("[%s] session=%d model=%s\n%s\n\n", m.Role, m.SessionID, m.Model, trim(m.Content, 500))
	}
	return nil
}
func runTransfer(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("transfer requires a target model")
	}
	c, s, err := loadStore()
	if err != nil {
		return err
	}
	targetModel := strings.TrimSpace(args[0])
	fmt.Println("Checking model", targetModel+"...")
	svc := transfer.NewService(c.OllamaURL, s)
	result, err := svc.Transfer(targetModel)
	if err != nil {
		return err
	}
	fmt.Println("Building context pack...")
	fmt.Println("Transferring session...")
	fmt.Println(targetModel, "is caught up.")
	fmt.Println()
	fmt.Println(result.Acknowledgement)
	return nil
}
func runCatchUp() error {
	_, s, err := loadStore()
	if err != nil {
		return err
	}
	msgs, err := s.RecentMessages(30)
	if err != nil {
		return err
	}
	fmt.Println("# llama-nest catch-up brief")
	for _, m := range msgs {
		fmt.Printf("- %s: %s\n", m.Role, trim(m.Content, 350))
	}
	return nil
}
func runDoctor() error {
	c, _, err := loadStore()
	if err != nil {
		return err
	}
	resp, err := http.Get(c.OllamaURL + "/api/tags")
	if err != nil {
		fmt.Println("ollama: not reachable -", err)
		return nil
	}
	defer resp.Body.Close()
	fmt.Println("ollama: reachable", resp.Status)
	fmt.Println("llama-nest: ok")
	return nil
}
func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
func parseChatAnswer(body []byte) string {
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`

		Response string `json:"response"`

		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &out); err == nil {
		if out.Message.Content != "" {
			return out.Message.Content
		}

		if out.Response != "" {
			return out.Response
		}

		if len(out.Choices) > 0 {
			return out.Choices[0].Message.Content
		}
	}

	return string(body)
}
func runWipe(args []string) error {
	confirmed := false

	for _, arg := range args {
		if arg == "--yes" || arg == "-y" {
			confirmed = true
		}
	}

	if !confirmed {
		fmt.Println("This will delete all captured llama-nest sessions, messages, transfers, and usage data.")
		fmt.Println()
		fmt.Println("Re-run with:")
		fmt.Println("  llama-nest wipe --yes")
		return nil
	}

	_, s, err := loadStore()
	if err != nil {
		return err
	}

	if err := s.Wipe(); err != nil {
		return err
	}

	fmt.Println("Deleted captured llama-nest context.")
	return nil
}
