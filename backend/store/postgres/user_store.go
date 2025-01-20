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
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt

	return s.loader.exec(ctx, CreateUserQuery,
		user.ID, user.Username, user.Password, user.Nickname,
		user.AvatarURL, user.CreatedAt, user.UpdatedAt)
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := s.loader.queryRow(ctx, GetUserByIDQuery,
		func(row pgx.Row) error {
			return row.Scan(&user.ID, &user.Username, &user.Password,
				&user.Nickname, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
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
				&user.Nickname, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
		},
		username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()
	return s.loader.exec(ctx, UpdateUserQuery,
		user.ID, user.Username, user.Nickname, user.AvatarURL, user.UpdatedAt)
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
