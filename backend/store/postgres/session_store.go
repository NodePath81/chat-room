package postgres

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateSession(ctx context.Context, session *models.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}

	tx, err := s.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create the session
	err = tx.(*Tx).loader.exec(ctx, CreateSessionQuery,
		session.ID, session.Name, session.CreatorID, session.CreatedAt)
	if err != nil {
		return err
	}

	// Add creator to the session with "creator" role
	err = tx.(*Tx).loader.exec(ctx, AddUserToSessionQuery,
		session.CreatorID, session.ID, "creator", session.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	session := &models.Session{}
	err := s.loader.queryRow(ctx, GetSessionsByIDQuery,
		func(row pgx.Row) error {
			return row.Scan(&session.ID, &session.Name, &session.CreatorID,
				&session.CreatedAt)
		},
		id)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Store) UpdateSession(ctx context.Context, session *models.Session) error {
	return s.loader.exec(ctx, UpdateSessionQuery,
		session.ID, session.Name)
}

func (s *Store) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.loader.exec(ctx, DeleteSessionQuery, id)
}

func (s *Store) GetSessionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	err := s.loader.queryRows(ctx, GetSessionsByIDsQuery,
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
		ids)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
