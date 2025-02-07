package postgres

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateMessage(ctx context.Context, message *models.Message) error {
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now().UTC()
	}

	return s.loader.exec(ctx, CreateMessageQuery,
		message.ID, message.Type, message.Content, message.UserID,
		message.SessionID, message.Timestamp)
}

func (s *Store) GetMessagesBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, before time.Time) ([]*models.Message, error) {
	var messages []*models.Message
	err := s.loader.queryRows(ctx, GetMessagesBySessionQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				msg := &models.Message{}
				err := rows.Scan(
					&msg.ID, &msg.Type, &msg.Content, &msg.UserID,
					&msg.SessionID, &msg.Timestamp,
				)
				if err != nil {
					return err
				}
				messages = append(messages, msg)
			}
			return nil
		},
		sessionID, before, limit)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *Store) GetMessageIDsBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, before time.Time) ([]uuid.UUID, error) {
	var messageIDs []uuid.UUID
	err := s.loader.queryRows(ctx, GetMessageIDsBySessionQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var id uuid.UUID
				if err := rows.Scan(&id); err != nil {
					return err
				}
				messageIDs = append(messageIDs, id)
			}
			return nil
		},
		sessionID, before, limit)
	if err != nil {
		return nil, err
	}
	return messageIDs, nil
}

func (s *Store) GetMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	msg := &models.Message{}
	err := s.loader.queryRow(ctx, GetMessageByIDQuery,
		func(row pgx.Row) error {
			return row.Scan(
				&msg.ID, &msg.Type, &msg.Content, &msg.UserID,
				&msg.SessionID, &msg.Timestamp,
			)
		},
		id)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *Store) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return s.loader.exec(ctx, DeleteMessageQuery, id)
}
