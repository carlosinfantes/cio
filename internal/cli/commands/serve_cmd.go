// Package commands implements the serve command for the API server.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/carlosinfantes/cio/internal/api"
	"github.com/carlosinfantes/cio/internal/cli/output"
	"github.com/carlosinfantes/cio/internal/config"
)

var (
	serveAddr string
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long: `Start the HTTP API server for frontend integration.

This enables building React, Vue, or other frontend applications
that communicate with the CIO - Chief Intelligence Officer engine.

Example:
  cio serve
  cio serve --port 8080
  cio serve --addr 0.0.0.0:8080

API Endpoints:
  POST /api/v1/session           Create a new chat session
  GET  /api/v1/session           List active sessions
  GET  /api/v1/session/{id}      Get session details
  DELETE /api/v1/session/{id}    End a session

  POST /api/v1/chat/{id}/message Send a message to Jordan

  GET  /api/v1/context           Get all CRF entities
  POST /api/v1/context           Create a CRF entity
  GET  /api/v1/context/{id}      Get a specific entity

  GET  /api/v1/decisions         List decisions (with filters)
  GET  /api/v1/decisions/{id}    Get a specific decision
  PATCH /api/v1/decisions/{id}/status  Update decision status

  POST /api/v1/panel/ask         Direct panel query (skip Jordan)`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", "", "Address to bind (overrides --port)")
	serveCmd.Flags().IntVar(&servePort, "port", 8765, "Port to listen on")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Loading config: %v", err))
		return err
	}

	if cfg.APIKey == "" {
		output.PrintError("No API key configured. Run: cio init")
		return fmt.Errorf("no API key")
	}

	// Determine address
	addr := serveAddr
	if addr == "" {
		addr = fmt.Sprintf(":%d", servePort)
	}

	// Create and start server
	server, err := api.NewServer(addr, cfg.APIKey, cfg.Model)
	if err != nil {
		output.PrintError(fmt.Sprintf("Creating server: %v", err))
		return err
	}

	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────┐")
	fmt.Println("│  💭 CIO - Chief Intelligence Officer API Server                        │")
	fmt.Println("├──────────────────────────────────────────────────────────┤")
	fmt.Printf("│  Listening on: http://localhost%s                   │\n", addr)
	fmt.Println("│                                                          │")
	fmt.Println("│  Endpoints:                                              │")
	fmt.Println("│    POST /api/v1/session       - Create session           │")
	fmt.Println("│    POST /api/v1/chat/{id}/message - Send message         │")
	fmt.Println("│    GET  /api/v1/context       - Get context              │")
	fmt.Println("│    GET  /api/v1/decisions     - List decisions           │")
	fmt.Println("│                                                          │")
	fmt.Println("│  Press Ctrl+C to stop                                    │")
	fmt.Println("└──────────────────────────────────────────────────────────┘")
	fmt.Println()

	return server.Start()
}
