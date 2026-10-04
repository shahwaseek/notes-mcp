// Command notes-mcp starts an MCP server over stdio for managing notes and
// tasks in a local SQLite database.
//
// All log output goes to stderr — stdout is reserved for MCP JSON-RPC messages.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/atiq/notes-mcp/internal/server"
	"github.com/atiq/notes-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// --- Flags ---
	dbPath := flag.String("db", defaultDBPath(), "Path to the SQLite database file")
	flag.Parse()

	// --- Logger (stderr only!) ---
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting notes-mcp server", "db", *dbPath)

	// --- Storage ---
	st, err := store.New(*dbPath, logger)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	// --- MCP Server ---
	srv := server.New(st)

	// --- Run over stdio ---
	logger.Info("server ready, listening on stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// defaultDBPath returns a sensible default location for the database.
// On most systems this is ~/.notes-mcp/notes.db.
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "notes.db"
	}
	return filepath.Join(home, ".notes-mcp", "notes.db")
}

func init() {
	// Safety check: if stdout is accidentally used for logging, catch it early.
	// This is a no-op guard — the real protection is using slog with stderr.
	_ = fmt.Sprintf // import fmt but don't use Println on stdout
}
