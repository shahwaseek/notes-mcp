# notes-mcp

An MCP (Model Context Protocol) server in Go that lets AI assistants (Claude Desktop, Claude Code, etc.) manage your **notes** and **tasks** stored in a local SQLite database.

## Features

- **8 tools** for notes and tasks management
- **Full-text search** via SQLite FTS5
- **Weekly summary** with overdue detection
- **MCP resource** exposing recent notes
- **Prompt template** for weekly planning
- **Pure Go** — no C compiler needed (uses `modernc.org/sqlite`)
- **Stdio transport** — works with any MCP client

## Tools

| Tool | Description |
|---|---|
| `add_note` | Save a note with title, body, and tags |
| `get_note` | Retrieve a note by ID |
| `search_notes` | Full-text search across all notes |
| `delete_note` | Delete a note by ID |
| `add_task` | Create a task with title, due date, priority |
| `list_tasks` | List tasks filtered by status and/or due date |
| `complete_task` | Mark a task as done |
| `weekly_summary` | Get counts of done, pending, overdue tasks |

## Quick Start

### Build

```bash
cd notes-mcp
go mod tidy
go build -o notes-mcp.exe ./cmd/notes-mcp
```

### Run directly

```bash
./notes-mcp --db ./my-notes.db
```

### Connect to Claude Code

```bash
claude mcp add notes -- /path/to/notes-mcp
```

### Connect to Claude Desktop

Add to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "notes": {
      "command": "/path/to/notes-mcp",
      "args": ["--db", "/path/to/notes.db"]
    }
  }
}
```

### Run with Docker

```bash
docker build -t notes-mcp .
docker run -i --rm -v notes-data:/data notes-mcp
```

## Architecture

```
cmd/notes-mcp/main.go       → Entrypoint (flags, logging, stdio)
internal/
  store/                     → SQLite persistence (no MCP dependency)
    store.go                 → DB init, migrations
    notes.go                 → Notes CRUD + FTS5 search
    tasks.go                 → Tasks CRUD + weekly summary
    migrations/001_init.sql  → Schema + FTS triggers
  tools/                     → MCP tool handlers
    notes.go                 → add/get/search/delete note
    tasks.go                 → add/list/complete task + summary
  server/server.go           → Wires tools, resources, prompts
```

The database layer is fully independent of MCP, so you can unit-test all logic with in-memory SQLite and no protocol overhead.

## Testing

```bash
go test ./...
```

Tests use in-memory SQLite (`:memory:`) — no files created on disk.

## Database

Default location: `~/.notes-mcp/notes.db`

Override with `--db /your/path.db`.

### Schema

**notes**: id, title, body, tags, created_at, updated_at  
**tasks**: id, title, description, status, due_date, priority, created_at, completed_at  
**notes_fts**: FTS5 virtual table for full-text search on notes

## License

MIT
