package store_test

import (
	"testing"

	"github.com/atiq/notes-mcp/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(":memory:", nil)
	if err != nil {
		t.Fatalf("creating test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// ---------------------------------------------------------------------------
// Notes
// ---------------------------------------------------------------------------

func TestAddNote(t *testing.T) {
	s := newTestStore(t)

	n, err := s.AddNote("Test Title", "Test Body", "go,mcp")
	if err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	if n.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if n.Title != "Test Title" {
		t.Errorf("title = %q, want %q", n.Title, "Test Title")
	}
	if n.Tags != "go,mcp" {
		t.Errorf("tags = %q, want %q", n.Tags, "go,mcp")
	}
}

func TestAddNote_EmptyTitle(t *testing.T) {
	s := newTestStore(t)
	_, err := s.AddNote("", "body", "")
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestGetNote(t *testing.T) {
	s := newTestStore(t)
	created, _ := s.AddNote("Fetch Me", "body here", "tag")

	got, err := s.GetNote(created.ID)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if got.Title != "Fetch Me" {
		t.Errorf("title = %q, want %q", got.Title, "Fetch Me")
	}
}

func TestGetNote_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetNote(999)
	if err == nil {
		t.Fatal("expected error for missing note")
	}
}

func TestSearchNotes(t *testing.T) {
	s := newTestStore(t)
	s.AddNote("Go Programming", "Learn Go basics", "go,tutorial")
	s.AddNote("Python Guide", "Python is great", "python")
	s.AddNote("Advanced Go", "Concurrency in Go", "go,advanced")

	results, err := s.SearchNotes("Go", 10)
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
	}
}

func TestSearchNotes_EmptyQuery(t *testing.T) {
	s := newTestStore(t)
	_, err := s.SearchNotes("", 10)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestDeleteNote(t *testing.T) {
	s := newTestStore(t)
	n, _ := s.AddNote("Delete Me", "bye", "")

	err := s.DeleteNote(n.ID)
	if err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}

	_, err = s.GetNote(n.ID)
	if err == nil {
		t.Fatal("expected note to be deleted")
	}
}

func TestDeleteNote_NotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.DeleteNote(999)
	if err == nil {
		t.Fatal("expected error for missing note")
	}
}

func TestListRecentNotes(t *testing.T) {
	s := newTestStore(t)
	s.AddNote("First", "a", "")
	s.AddNote("Second", "b", "")
	s.AddNote("Third", "c", "")

	notes, err := s.ListRecentNotes(2)
	if err != nil {
		t.Fatalf("ListRecentNotes: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
	// Most recent first.
	if notes[0].Title != "Third" {
		t.Errorf("first note title = %q, want %q", notes[0].Title, "Third")
	}
}

// ---------------------------------------------------------------------------
// Tasks
// ---------------------------------------------------------------------------

func TestAddTask(t *testing.T) {
	s := newTestStore(t)

	task, err := s.AddTask("Ship feature", "Deploy to prod", "2025-12-31", 3)
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if task.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if task.Status != "todo" {
		t.Errorf("status = %q, want %q", task.Status, "todo")
	}
	if task.DueDate != "2025-12-31" {
		t.Errorf("due_date = %q, want %q", task.DueDate, "2025-12-31")
	}
}

func TestAddTask_EmptyTitle(t *testing.T) {
	s := newTestStore(t)
	_, err := s.AddTask("", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestAddTask_BadDueDate(t *testing.T) {
	s := newTestStore(t)
	_, err := s.AddTask("task", "", "not-a-date", 0)
	if err == nil {
		t.Fatal("expected error for invalid due date")
	}
}

func TestListTasks(t *testing.T) {
	s := newTestStore(t)
	s.AddTask("Task A", "", "2025-01-01", 1)
	s.AddTask("Task B", "", "2025-06-01", 2)
	s.AddTask("Task C", "", "", 0)

	// All tasks.
	tasks, err := s.ListTasks("", "")
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}

	// Filter by status.
	tasks, err = s.ListTasks("todo", "")
	if err != nil {
		t.Fatalf("ListTasks(todo): %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 todo tasks, got %d", len(tasks))
	}

	// Filter by due_before.
	tasks, err = s.ListTasks("", "2025-03-01")
	if err != nil {
		t.Fatalf("ListTasks(due_before): %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task due before 2025-03-01, got %d", len(tasks))
	}
}

func TestListTasks_InvalidStatus(t *testing.T) {
	s := newTestStore(t)
	_, err := s.ListTasks("invalid", "")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCompleteTask(t *testing.T) {
	s := newTestStore(t)
	task, _ := s.AddTask("Finish tests", "", "", 0)

	completed, err := s.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	if completed.Status != "done" {
		t.Errorf("status = %q, want %q", completed.Status, "done")
	}
	if completed.CompletedAt == "" {
		t.Error("expected completed_at to be set")
	}
}

func TestCompleteTask_AlreadyDone(t *testing.T) {
	s := newTestStore(t)
	task, _ := s.AddTask("Already done", "", "", 0)
	s.CompleteTask(task.ID)

	_, err := s.CompleteTask(task.ID)
	if err == nil {
		t.Fatal("expected error for already completed task")
	}
}

func TestCompleteTask_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.CompleteTask(999)
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestGetWeeklySummary(t *testing.T) {
	s := newTestStore(t)
	s.AddTask("Pending 1", "", "", 0)
	s.AddTask("Pending 2", "", "", 0)
	task, _ := s.AddTask("Done task", "", "", 0)
	s.CompleteTask(task.ID)

	summary, err := s.GetWeeklySummary()
	if err != nil {
		t.Fatalf("GetWeeklySummary: %v", err)
	}
	if summary.TotalTasks != 3 {
		t.Errorf("total = %d, want 3", summary.TotalTasks)
	}
	if summary.DoneTasks != 1 {
		t.Errorf("done = %d, want 1", summary.DoneTasks)
	}
	if summary.PendingTasks != 2 {
		t.Errorf("pending = %d, want 2", summary.PendingTasks)
	}
}
