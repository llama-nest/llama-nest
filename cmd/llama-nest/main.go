package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/riteshmishra/llama-nest/internal/api"
	"github.com/riteshmishra/llama-nest/internal/config"
	"github.com/riteshmishra/llama-nest/internal/db"
	"github.com/riteshmishra/llama-nest/internal/proxy"
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
	case "status":
		err = runStatus()
	case "search":
		err = runSearch(os.Args[2:])
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
  init        create config and local store
  start       start API server and Ollama proxy
  status      print local status
  search Q    search captured context
  catch-up    print recent context brief
  doctor      check Ollama connectivity`)
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
