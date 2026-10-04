// Package tools defines MCP tool handlers for notes.
// Each tool is a typed struct (input) with a handler function that the MCP
// server's generic AddTool wires up automatically.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atiq/notes-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// add_note
// ---------------------------------------------------------------------------

// AddNoteInput defines the schema for the add_note tool.
type AddNoteInput struct {
	Title string `json:"title"          jsonschema:"Title of the note"`
	Body  string `json:"body,omitempty" jsonschema:"Body/content of the note"`
	Tags  string `json:"tags,omitempty" jsonschema:"Comma-separated tags for the note"`
}

// AddNoteHandler returns a handler that saves a note.
func AddNoteHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input AddNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddNoteInput) (*mcp.CallToolResult, any, error) {
		note, err := s.AddNote(input.Title, input.Body, input.Tags)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(note, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Note created successfully:\n%s", string(data))}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// get_note
// ---------------------------------------------------------------------------

// GetNoteInput defines the schema for the get_note tool.
type GetNoteInput struct {
	ID int64 `json:"id" jsonschema:"ID of the note to retrieve"`
}

// GetNoteHandler returns a handler that retrieves a single note.
func GetNoteHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input GetNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetNoteInput) (*mcp.CallToolResult, any, error) {
		note, err := s.GetNote(input.ID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(note, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// search_notes
// ---------------------------------------------------------------------------

// SearchNotesInput defines the schema for the search_notes tool.
type SearchNotesInput struct {
	Query string `json:"query"           jsonschema:"Full-text search query across note titles, bodies, and tags"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 20)"`
}

// SearchNotesHandler returns a handler for full-text search.
func SearchNotesHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input SearchNotesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchNotesInput) (*mcp.CallToolResult, any, error) {
		notes, err := s.SearchNotes(input.Query, input.Limit)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		if len(notes) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No notes found matching your query."}},
			}, nil, nil
		}
		data, _ := json.MarshalIndent(notes, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Found %d note(s):\n%s", len(notes), string(data))}},
		}, nil, nil
	}
}

// ---------------------------------------------------------------------------
// delete_note
// ---------------------------------------------------------------------------

// DeleteNoteInput defines the schema for the delete_note tool.
type DeleteNoteInput struct {
	ID int64 `json:"id" jsonschema:"ID of the note to delete"`
}

// DeleteNoteHandler returns a handler that deletes a note.
func DeleteNoteHandler(s *store.Store) func(ctx context.Context, req *mcp.CallToolRequest, input DeleteNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteNoteInput) (*mcp.CallToolResult, any, error) {
		err := s.DeleteNote(input.ID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Note %d deleted successfully.", input.ID)}},
		}, nil, nil
	}
}
