package tools_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/atiq/notes-mcp/internal/store"
	"github.com/atiq/notes-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st, err := store.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func extractText(res *mcp.CallToolResult) string {
	if res == nil || len(res.Content) == 0 {
		return ""
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func TestNotesTools(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	addHandler := tools.AddNoteHandler(st)
	getHandler := tools.GetNoteHandler(st)
	searchHandler := tools.SearchNotesHandler(st)
	deleteHandler := tools.DeleteNoteHandler(st)

	// 1. Add Note
	addRes, _, err := addHandler(ctx, nil, tools.AddNoteInput{
		Title: "Meeting Notes",
		Body:  "Discussed architecture and MCP design",
		Tags:  "work,mcp",
	})
	if err != nil {
		t.Fatalf("unexpected error adding note: %v", err)
	}
	if addRes.IsError {
		t.Fatalf("expected successful note addition, got error: %s", extractText(addRes))
	}
	if !strings.Contains(extractText(addRes), "Meeting Notes") {
		t.Errorf("expected text to contain note title, got: %s", extractText(addRes))
	}

	// 2. Add Note - Validation failure
	addErrRes, _, _ := addHandler(ctx, nil, tools.AddNoteInput{
		Title: "   ",
		Body:  "empty title",
	})
	if !addErrRes.IsError {
		t.Errorf("expected error for empty note title, got success")
	}

	// 3. Get Note
	getRes, _, err := getHandler(ctx, nil, tools.GetNoteInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error getting note: %v", err)
	}
	if getRes.IsError {
		t.Fatalf("expected success getting note 1, got: %s", extractText(getRes))
	}
	if !strings.Contains(extractText(getRes), "Discussed architecture") {
		t.Errorf("expected body in get result, got: %s", extractText(getRes))
	}

	// 4. Get Note - Not Found
	getNotFoundRes, _, _ := getHandler(ctx, nil, tools.GetNoteInput{ID: 999})
	if !getNotFoundRes.IsError {
		t.Errorf("expected error when getting non-existent note")
	}

	// 5. Search Notes
	searchRes, _, err := searchHandler(ctx, nil, tools.SearchNotesInput{Query: "architecture", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error searching notes: %v", err)
	}
	if searchRes.IsError {
		t.Fatalf("expected search success, got: %s", extractText(searchRes))
	}
	if !strings.Contains(extractText(searchRes), "Found 1 note(s)") {
		t.Errorf("expected 1 note match, got: %s", extractText(searchRes))
	}

	// 6. Search Notes - No Results
	searchEmptyRes, _, _ := searchHandler(ctx, nil, tools.SearchNotesInput{Query: "nonexistentkeyword"})
	if !strings.Contains(extractText(searchEmptyRes), "No notes found") {
		t.Errorf("expected 'No notes found', got: %s", extractText(searchEmptyRes))
	}

	// 7. Delete Note
	delRes, _, err := deleteHandler(ctx, nil, tools.DeleteNoteInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error deleting note: %v", err)
	}
	if delRes.IsError {
		t.Fatalf("expected success deleting note, got: %s", extractText(delRes))
	}

	// 8. Delete Note - Not Found
	delNotFoundRes, _, _ := deleteHandler(ctx, nil, tools.DeleteNoteInput{ID: 999})
	if !delNotFoundRes.IsError {
		t.Errorf("expected error deleting non-existent note")
	}
}

func TestTasksTools(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	addHandler := tools.AddTaskHandler(st)
	listHandler := tools.ListTasksHandler(st)
	completeHandler := tools.CompleteTaskHandler(st)
	summaryHandler := tools.WeeklySummaryHandler(st)

	// 1. Add Task
	addRes, _, err := addHandler(ctx, nil, tools.AddTaskInput{
		Title:       "Deploy server",
		Description: "Run docker container",
		DueDate:     "2026-10-10",
		Priority:    2,
	})
	if err != nil {
		t.Fatalf("unexpected error adding task: %v", err)
	}
	if addRes.IsError {
		t.Fatalf("expected success adding task, got: %s", extractText(addRes))
	}

	// 2. Add Task - Invalid
	addErrRes, _, _ := addHandler(ctx, nil, tools.AddTaskInput{
		Title: "",
	})
	if !addErrRes.IsError {
		t.Errorf("expected error for empty task title")
	}

	// 3. List Tasks
	listRes, _, err := listHandler(ctx, nil, tools.ListTasksInput{Status: "todo"})
	if err != nil {
		t.Fatalf("unexpected error listing tasks: %v", err)
	}
	if listRes.IsError {
		t.Fatalf("expected list tasks success, got: %s", extractText(listRes))
	}
	if !strings.Contains(extractText(listRes), "Deploy server") {
		t.Errorf("expected 'Deploy server' in list, got: %s", extractText(listRes))
	}

	// 4. Complete Task
	compRes, _, err := completeHandler(ctx, nil, tools.CompleteTaskInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error completing task: %v", err)
	}
	if compRes.IsError {
		t.Fatalf("expected success completing task, got: %s", extractText(compRes))
	}

	// 5. Complete Task - Already Completed
	compErrRes, _, _ := completeHandler(ctx, nil, tools.CompleteTaskInput{ID: 1})
	if !compErrRes.IsError {
		t.Errorf("expected error completing already-completed task")
	}

	// 6. Weekly Summary
	sumRes, _, err := summaryHandler(ctx, nil, tools.WeeklySummaryInput{})
	if err != nil {
		t.Fatalf("unexpected error getting weekly summary: %v", err)
	}
	if sumRes.IsError {
		t.Fatalf("expected summary success, got: %s", extractText(sumRes))
	}
	if !strings.Contains(extractText(sumRes), "Weekly Summary") {
		t.Errorf("expected weekly summary text, got: %s", extractText(sumRes))
	}
}
