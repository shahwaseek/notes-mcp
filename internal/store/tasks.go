package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Task represents a single task in the database.
type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "todo" or "done"
	DueDate     string `json:"due_date,omitempty"`
	Priority    int    `json:"priority"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// AddTask creates a new task and returns it with its generated ID.
func (s *Store) AddTask(title, description, dueDate string, priority int) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}

	// Validate due_date format if provided.
	if dueDate != "" {
		if _, err := time.Parse("2006-01-02", dueDate); err != nil {
			return nil, fmt.Errorf("due_date must be in YYYY-MM-DD format, got %q", dueDate)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var dueDatePtr *string
	if dueDate != "" {
		dueDatePtr = &dueDate
	}

	result, err := s.db.Exec(
		`INSERT INTO tasks (title, description, status, due_date, priority, created_at)
		 VALUES (?, ?, 'todo', ?, ?, ?)`,
		title, description, dueDatePtr, priority, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	s.logger.Info("task created", "id", id, "title", title)
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      "todo",
		DueDate:     dueDate,
		Priority:    priority,
		CreatedAt:   now,
	}, nil
}

// ListTasks returns tasks filtered by optional status and due_before date.
// Pass empty strings to skip filters.
func (s *Store) ListTasks(status, dueBefore string) ([]Task, error) {
	query := `SELECT id, title, description, status, COALESCE(due_date, ''), priority, created_at, COALESCE(completed_at, '') FROM tasks WHERE 1=1`
	var args []any

	if status != "" {
		if status != "todo" && status != "done" {
			return nil, fmt.Errorf("status must be 'todo' or 'done', got %q", status)
		}
		query += ` AND status = ?`
		args = append(args, status)
	}
	if dueBefore != "" {
		if _, err := time.Parse("2006-01-02", dueBefore); err != nil {
			return nil, fmt.Errorf("due_before must be in YYYY-MM-DD format, got %q", dueBefore)
		}
		query += ` AND due_date IS NOT NULL AND due_date <= ?`
		args = append(args, dueBefore)
	}

	query += ` ORDER BY priority DESC, due_date ASC, created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status,
			&t.DueDate, &t.Priority, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// CompleteTask marks a task as done by ID. Returns an error if the task
// does not exist or is already completed.
func (s *Store) CompleteTask(id int64) (*Task, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.Exec(
		`UPDATE tasks SET status = 'done', completed_at = ? WHERE id = ? AND status = 'todo'`,
		now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("completing task: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		// Check whether the task exists at all.
		var status string
		err := s.db.QueryRow(`SELECT status FROM tasks WHERE id = ?`, id).Scan(&status)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %d not found", id)
		}
		if err != nil {
			return nil, fmt.Errorf("querying task: %w", err)
		}
		return nil, fmt.Errorf("task %d is already done", id)
	}

	s.logger.Info("task completed", "id", id)

	// Return the updated task.
	t := &Task{}
	err = s.db.QueryRow(
		`SELECT id, title, description, status, COALESCE(due_date, ''), priority, created_at, COALESCE(completed_at, '') FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.Priority, &t.CreatedAt, &t.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("fetching completed task: %w", err)
	}
	return t, nil
}

// WeeklySummary holds aggregate counts for the weekly_summary tool.
type WeeklySummary struct {
	TotalTasks     int `json:"total_tasks"`
	DoneTasks      int `json:"done_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	OverdueTasks   int `json:"overdue_tasks"`
	DoneThisWeek   int `json:"done_this_week"`
	DueThisWeek    int `json:"due_this_week"`
}

// GetWeeklySummary returns counts of done, pending, and overdue tasks.
func (s *Store) GetWeeklySummary() (*WeeklySummary, error) {
	today := time.Now().UTC().Format("2006-01-02")

	// Calculate the start of the current week (Monday).
	now := time.Now().UTC()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -int(weekday-time.Monday)).Format("2006-01-02")
	weekEnd := now.AddDate(0, 0, int(7-weekday)).Format("2006-01-02")

	summary := &WeeklySummary{}

	// Total and status counts.
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&summary.TotalTasks)
	if err != nil {
		return nil, fmt.Errorf("counting total tasks: %w", err)
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'done'`).Scan(&summary.DoneTasks)
	if err != nil {
		return nil, fmt.Errorf("counting done tasks: %w", err)
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'todo'`).Scan(&summary.PendingTasks)
	if err != nil {
		return nil, fmt.Errorf("counting pending tasks: %w", err)
	}

	// Overdue: todo tasks whose due_date is strictly before today.
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'todo' AND due_date IS NOT NULL AND due_date < ?`, today).Scan(&summary.OverdueTasks)
	if err != nil {
		return nil, fmt.Errorf("counting overdue tasks: %w", err)
	}

	// Done this week.
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'done' AND completed_at >= ?`, weekStart).Scan(&summary.DoneThisWeek)
	if err != nil {
		return nil, fmt.Errorf("counting done this week: %w", err)
	}

	// Due this week (todo only).
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'todo' AND due_date IS NOT NULL AND due_date >= ? AND due_date <= ?`, weekStart, weekEnd).Scan(&summary.DueThisWeek)
	if err != nil {
		return nil, fmt.Errorf("counting due this week: %w", err)
	}

	return summary, nil
}
