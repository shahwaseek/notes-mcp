// Package server wires MCP tools, resources, and prompts into an mcp.Server.
package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atiq/notes-mcp/internal/store"
	"github.com/atiq/notes-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "notes-mcp"
	serverVersion = "0.1.0"
)

// New creates a fully-configured MCP server with all tools, resources, and
// prompts registered.
func New(st *store.Store) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: serverVersion},
		nil,
	)

	registerTools(s, st)
	registerResources(s, st)
	registerPrompts(s, st)

	return s
}

// registerTools registers all note and task tools with the MCP server.
func registerTools(s *mcp.Server, st *store.Store) {
	// --- Notes ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_note",
		Description: "Save a new note with a title, body, and optional comma-separated tags. Returns the created note with its ID.",
	}, tools.AddNoteHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_note",
		Description: "Retrieve a single note by its numeric ID. Returns the full note including title, body, tags, and timestamps.",
	}, tools.GetNoteHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_notes",
		Description: "Full-text search across note titles, bodies, and tags. Uses SQLite FTS5 for fast, ranked results. Supports prefix queries (e.g. 'prog*') and phrase matching ('\"exact phrase\"').",
	}, tools.SearchNotesHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_note",
		Description: "Permanently delete a note by its numeric ID. This action cannot be undone.",
	}, tools.DeleteNoteHandler(st))

	// --- Tasks ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_task",
		Description: "Create a new task. Provide a title (required), optional description, due date (YYYY-MM-DD), and priority (higher number = more important, default 0).",
	}, tools.AddTaskHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List tasks with optional filters. Filter by status ('todo' or 'done') and/or due_before date (YYYY-MM-DD). Results are sorted by priority (descending), then due date.",
	}, tools.ListTasksHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "complete_task",
		Description: "Mark a task as done by its numeric ID. Sets the status to 'done' and records the completion timestamp. Returns an error if the task is already completed.",
	}, tools.CompleteTaskHandler(st))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "weekly_summary",
		Description: "Get a summary of your task progress: total tasks, completed, pending, overdue, and counts for the current week. Takes no input.",
	}, tools.WeeklySummaryHandler(st))
}

// registerResources exposes recent notes as an MCP resource.
func registerResources(s *mcp.Server, st *store.Store) {
	s.AddResource(
		&mcp.Resource{
			URI:         "notes://recent",
			Name:        "Recent Notes",
			Description: "The 10 most recently created notes.",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			notes, err := st.ListRecentNotes(10)
			if err != nil {
				return nil, fmt.Errorf("listing recent notes: %w", err)
			}
			data, _ := json.MarshalIndent(notes, "", "  ")
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      "notes://recent",
					MIMEType: "application/json",
					Text:     string(data),
				}},
			}, nil
		},
	)
}

// registerPrompts adds prompt templates to the server.
func registerPrompts(s *mcp.Server, st *store.Store) {
	s.AddPrompt(
		&mcp.Prompt{
			Name:        "plan_my_week",
			Description: "Generate a weekly plan based on current pending tasks and recent notes. Provides context about overdue items and upcoming due dates.",
		},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			// Gather context: pending tasks.
			pendingTasks, err := st.ListTasks("todo", "")
			if err != nil {
				return nil, fmt.Errorf("listing pending tasks: %w", err)
			}
			tasksJSON, _ := json.MarshalIndent(pendingTasks, "", "  ")

			// Recent notes for context.
			notes, err := st.ListRecentNotes(5)
			if err != nil {
				return nil, fmt.Errorf("listing recent notes: %w", err)
			}
			notesJSON, _ := json.MarshalIndent(notes, "", "  ")

			// Weekly summary.
			summary, err := st.GetWeeklySummary()
			if err != nil {
				return nil, fmt.Errorf("getting weekly summary: %w", err)
			}
			summaryJSON, _ := json.MarshalIndent(summary, "", "  ")

			promptText := fmt.Sprintf(`Here is context about my current tasks and notes. Please help me plan my week.

## Current Task Summary
%s

## Pending Tasks
%s

## Recent Notes (for context)
%s

Based on this information, please:
1. Identify which tasks are overdue and need immediate attention
2. Suggest a day-by-day plan for the week
3. Flag any tasks that might need to be broken down further
4. Suggest priorities based on due dates and current priority levels`,
				string(summaryJSON), string(tasksJSON), string(notesJSON))

			return &mcp.GetPromptResult{
				Description: "Plan your week based on current tasks and notes",
				Messages: []*mcp.PromptMessage{
					{
						Role:    "user",
						Content: &mcp.TextContent{Text: promptText},
					},
				},
			}, nil
		},
	)
}
