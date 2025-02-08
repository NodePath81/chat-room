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

func (s *Store) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	var userSessions []*models.UserSession
	err := s.loader.queryRows(ctx, GetUserSessionsQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				session := &models.Session{}
				userSession := &models.UserSession{
					UserID: userID,
				}
				err := rows.Scan(
					&session.ID, &session.Name, &session.CreatorID,
					&session.CreatedAt, &userSession.Role, &userSession.JoinedAt,
				)
				if err != nil {
					return err
				}
				userSession.SessionID = session.ID
				userSessions = append(userSessions, userSession)
			}
			return nil
		},
		userID)
	if err != nil {
		return nil, err
	}
	return userSessions, nil
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

func (s *Store) GetSessionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var sessionIDs []uuid.UUID
	err := s.loader.queryRows(ctx, GetSessionIDsByUserIDQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var id uuid.UUID
				if err := rows.Scan(&id); err != nil {
					return err
				}
				sessionIDs = append(sessionIDs, id)
			}
			return nil
		},
		userID)
	if err != nil {
		return nil, err
	}
	return sessionIDs, nil
}

func (s *Store) GetUserIDsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := s.loader.queryRows(ctx, GetUserIDsBySessionIDQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var id uuid.UUID
				if err := rows.Scan(&id); err != nil {
					return err
				}
				userIDs = append(userIDs, id)
			}
			return nil
		},
		sessionID)
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (s *Store) GetUserSessionsBySessionIDAndUserIDs(ctx context.Context, sessionID uuid.UUID, userIDs []uuid.UUID) ([]*models.UserSession, error) {
	var userSessions []*models.UserSession
	err := s.loader.queryRows(ctx, GetUserSessionsBySessionIDAndUserIDsQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				userSession := &models.UserSession{}
				err := rows.Scan(
					&userSession.UserID,
					&userSession.SessionID,
					&userSession.Role,
					&userSession.JoinedAt,
				)
				if err != nil {
					return err
				}
				userSessions = append(userSessions, userSession)
			}
			return nil
		},
		sessionID, userIDs)
	if err != nil {
		return nil, err
	}
	return userSessions, nil
}
