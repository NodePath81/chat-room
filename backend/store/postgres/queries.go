package postgres

import (
	"embed"
	"fmt"
	"strings"
)

// QueryName represents the name of a SQL query
type QueryName string

// Define all query names as constants
const (
	// User queries
	CreateUserQuery          QueryName = "CreateUser"
	GetUserByIDQuery         QueryName = "GetUserByID"
	GetUserByUsernameQuery   QueryName = "GetUserByUsername"
	UpdateUserQuery          QueryName = "UpdateUser"
	DeleteUserQuery          QueryName = "DeleteUser"
	CheckUsernameExistsQuery QueryName = "CheckUsernameExists"
	CheckNicknameExistsQuery QueryName = "CheckNicknameExists"

	// Session queries
	CreateSessionQuery      QueryName = "CreateSession"
	GetSessionByIDQuery     QueryName = "GetSessionByID"
	GetUserSessionsQuery    QueryName = "GetUserSessions"
	UpdateSessionQuery      QueryName = "UpdateSession"
	DeleteSessionQuery      QueryName = "DeleteSession"
	AddUserToSessionQuery   QueryName = "AddUserToSession"
	RemoveUserFromSession   QueryName = "RemoveUserFromSession"
	GetSessionUsersQuery    QueryName = "GetSessionUsers"
	GetUserSessionRoleQuery QueryName = "GetUserSessionRole"

	// Message queries
	CreateMessageQuery        QueryName = "CreateMessage"
	GetMessagesBySessionQuery QueryName = "GetMessagesBySessionID"
	DeleteMessageQuery        QueryName = "DeleteMessage"
)

// queryStore holds all loaded SQL queries
type queryStore struct {
	queries map[QueryName]string
}

// newQueryStore creates a new query store and loads all SQL queries
func newQueryStore(fs embed.FS) (*queryStore, error) {
	qs := &queryStore{
		queries: make(map[QueryName]string),
	}

	// Load queries from each SQL file
	files := []string{
		"queries/users.sql",
		"queries/sessions.sql",
		"queries/messages.sql",
	}

	for _, file := range files {
		content, err := fs.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading SQL file %s: %w", file, err)
		}

		if err := qs.parseQueries(string(content)); err != nil {
			return nil, fmt.Errorf("parsing queries from %s: %w", file, err)
		}
	}

	return qs, nil
}

// parseQueries parses SQL queries from a file content
func (qs *queryStore) parseQueries(content string) error {
	queries := strings.Split(content, "\n\n")
	for _, query := range queries {
		if strings.TrimSpace(query) == "" {
			continue
		}

		lines := strings.Split(query, "\n")
		if len(lines) < 2 {
			continue
		}

		// Parse query name from comment
		nameLine := strings.TrimSpace(lines[0])
		if !strings.HasPrefix(nameLine, "-- name:") {
			continue
		}

		parts := strings.Fields(nameLine)
		if len(parts) < 3 {
			continue
		}

		name := QueryName(parts[2])
		queryText := strings.Join(lines[1:], "\n")
		qs.queries[name] = strings.TrimSpace(queryText)
	}

	return nil
}

// get returns the SQL query for the given name
func (qs *queryStore) get(name QueryName) (string, error) {
	query, ok := qs.queries[name]
	if !ok {
		return "", fmt.Errorf("query %q not found", name)
	}
	return query, nil
}
