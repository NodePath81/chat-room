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
	session.CreatedAt = time.Now().UTC()
	session.UpdatedAt = session.CreatedAt

	return s.loader.exec(ctx, CreateSessionQuery,
		session.ID, session.Name, session.CreatorID, session.CreatedAt, session.UpdatedAt)
}

func (s *Store) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	session := &models.Session{}
	err := s.loader.queryRow(ctx, GetSessionByIDQuery,
		func(row pgx.Row) error {
			return row.Scan(&session.ID, &session.Name, &session.CreatorID,
				&session.CreatedAt, &session.UpdatedAt)
		},
		id)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Store) UpdateSession(ctx context.Context, session *models.Session) error {
	session.UpdatedAt = time.Now().UTC()
	return s.loader.exec(ctx, UpdateSessionQuery,
		session.ID, session.Name, session.UpdatedAt)
}

func (s *Store) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.loader.exec(ctx, DeleteSessionQuery, id)
}
