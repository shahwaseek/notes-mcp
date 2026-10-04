package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atiq/notes-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// add_task
// ---------------------------------------------------------------------------

// AddTaskInput defines the schema for the add_task tool.
type AddTaskInput struct {
	Title       string `json:"title"                 jsonschema:"Title of the task"`
	Description string `json:"description,omitempty" jsonschema:"Detailed description of the task"`
	DueDate     string `json:"due_date,omitempty"    jsonschema:"Due date in YYYY-MM-DD format"`
	Priority    int    `json:"priority,omitempty"    jsonschema:"Priority level (higher number = higher priority; default 0)"`
}

// AddTaskHandler returns a handler that creates a new task.
func AddTaskHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input AddTaskInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddTaskInput) (*mcp.CallToolResult, any, error) {
		task, err := s.AddTask(input.Title, input.Description, input.DueDate, input.Priority)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(task, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Task created successfully:\n%s", string(data))}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// list_tasks
// ---------------------------------------------------------------------------

// ListTasksInput defines the schema for the list_tasks tool.
type ListTasksInput struct {
	Status    string `json:"status,omitempty"     jsonschema:"Filter by status: 'todo' or 'done'. Leave empty for all tasks."`
	DueBefore string `json:"due_before,omitempty" jsonschema:"Show tasks due on or before this date (YYYY-MM-DD)"`
}

// ListTasksHandler returns a handler that lists tasks with optional filters.
func ListTasksHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input ListTasksInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListTasksInput) (*mcp.CallToolResult, any, error) {
		tasks, err := s.ListTasks(input.Status, input.DueBefore)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		if len(tasks) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No tasks found."}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(tasks, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Found %d task(s):\n%s", len(tasks), string(data))}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// complete_task
// ---------------------------------------------------------------------------

// CompleteTaskInput defines the schema for the complete_task tool.
type CompleteTaskInput struct {
	ID int64 `json:"id" jsonschema:"ID of the task to mark as done"`
}

// CompleteTaskHandler returns a handler that marks a task as completed.
func CompleteTaskHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input CompleteTaskInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CompleteTaskInput) (*mcp.CallToolResult, any, error) {
		task, err := s.CompleteTask(input.ID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(task, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Task completed:\n%s", string(data))}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// weekly_summary
// ---------------------------------------------------------------------------

// WeeklySummaryInput is empty — the tool takes no input.
type WeeklySummaryInput struct{}

// WeeklySummaryHandler returns a handler that produces a weekly summary.
func WeeklySummaryHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input WeeklySummaryInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WeeklySummaryInput) (*mcp.CallToolResult, any, error) {
		summary, err := s.GetWeeklySummary()
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		text := fmt.Sprintf(`📊 Weekly Summary
━━━━━━━━━━━━━━━━━━━
Total tasks:      %d
✅ Done (all):     %d
📋 Pending:        %d
⚠️  Overdue:        %d
━━━━━━━━━━━━━━━━━━━
This week:
  Completed:      %d
  Still due:      %d`,
			summary.TotalTasks,
			summary.DoneTasks,
			summary.PendingTasks,
			summary.OverdueTasks,
			summary.DoneThisWeek,
			summary.DueThisWeek,
		)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	}
}
