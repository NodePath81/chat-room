package postgres

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) AddUserToSession(ctx context.Context, userID, sessionID uuid.UUID, role string) error {
	return s.loader.exec(ctx, AddUserToSessionQuery,
		userID, sessionID, role, time.Now().UTC())
}

func (s *Store) RemoveUserFromSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.loader.exec(ctx, RemoveUserFromSession,
		userID, sessionID)
}

func (s *Store) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	err := s.loader.queryRows(ctx, GetUserSessionsQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				session := &models.Session{}
				err := rows.Scan(
					&session.ID, &session.Name, &session.CreatorID,
					&session.CreatedAt,
				)
				if err != nil {
					return err
				}
				sessions = append(sessions, session)
			}
			return nil
		},
		userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) GetSessionUsers(ctx context.Context, sessionID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	err := s.loader.queryRows(ctx, GetSessionUsersQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				user := &models.User{}
				err := rows.Scan(
					&user.ID, &user.Username, &user.Nickname,
					&user.AvatarURL, &user.CreatedAt,
				)
				if err != nil {
					return err
				}
				users = append(users, user)
			}
			return nil
		},
		sessionID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) GetUserSessionRole(ctx context.Context, userID, sessionID uuid.UUID) (string, error) {
	var role string
	err := s.loader.queryRow(ctx, GetUserSessionRoleQuery,
		func(row pgx.Row) error {
			return row.Scan(&role)
		},
		userID, sessionID)
	if err != nil {
		return "", err
	}
	return role, nil
}
