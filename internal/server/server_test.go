package server_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/atiq/notes-mcp/internal/server"
	"github.com/atiq/notes-mcp/internal/store"
)

func TestNewServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st, err := store.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	defer st.Close()

	srv := server.New(st)
	if srv == nil {
		t.Fatal("expected non-nil server instance")
	}
}
