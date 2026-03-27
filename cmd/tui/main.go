package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AlexiaChen/issue-kanban-mcp/internal/apiclient"
	"github.com/AlexiaChen/issue-kanban-mcp/internal/tui"
)

func main() {
	server := flag.String("server", "http://localhost:9292", "server URL")
	flag.Parse()

	client := apiclient.New(*server)
	app := tui.NewApp(client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
