package postgres

import (
	"context"
	"time"

	"chat-room/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	return s.loader.exec(ctx, CreateUserQuery,
		user.ID, user.Username, user.Password, user.Nickname,
		user.AvatarURL, user.CreatedAt)
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := s.loader.queryRow(ctx, GetUserByIDQuery,
		func(row pgx.Row) error {
			return row.Scan(&user.ID, &user.Username, &user.Password,
				&user.Nickname, &user.AvatarURL, &user.CreatedAt)
		},
		id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := s.loader.queryRow(ctx, GetUserByUsernameQuery,
		func(row pgx.Row) error {
			return row.Scan(&user.ID, &user.Username, &user.Password,
				&user.Nickname, &user.AvatarURL, &user.CreatedAt)
		},
		username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, user *models.User) error {
	return s.loader.exec(ctx, UpdateUserQuery,
		user.ID, user.Username, user.Nickname, user.AvatarURL)
}

func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.loader.exec(ctx, DeleteUserQuery, id)
}

func (s *Store) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.loader.queryRow(ctx, CheckUsernameExistsQuery,
		func(row pgx.Row) error {
			return row.Scan(&exists)
		},
		username)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) CheckNicknameExists(ctx context.Context, nickname string) (bool, error) {
	var exists bool
	err := s.loader.queryRow(ctx, CheckNicknameExistsQuery,
		func(row pgx.Row) error {
			return row.Scan(&exists)
		},
		nickname)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	err := s.loader.queryRows(ctx, GetUsersByIDsQuery,
		func(rows pgx.Rows) error {
			for rows.Next() {
				user := &models.User{}
				err := rows.Scan(
					&user.ID, &user.Username, &user.Password,
					&user.Nickname, &user.AvatarURL, &user.CreatedAt,
				)
				if err != nil {
					return err
				}
				users = append(users, user)
			}
			return nil
		},
		ids)
	if err != nil {
		return nil, err
	}
	return users, nil
}
