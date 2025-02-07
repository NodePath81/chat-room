package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat-room/config"
	"chat-room/models"
	"chat-room/store"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CachedUser represents the user data that will be stored in Redis
type CachedUser struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
}

// CachedSession represents a session with its members
type CachedSession struct {
	Session *models.Session `json:"session"`
	Members []*CachedUser   `json:"members"`
}

const (
	// Cache keys
	userKey         = "user:%s"          // user:{userID}
	sessionKey      = "session:%s"       // session:{sessionID}
	messageKey      = "msg:%s"           // msg:{messageID}
	userSessionsKey = "user:%s:sessions" // user:{userID}:sessions

	// Cache expiration times
	userExpiration        = 30 * time.Minute
	sessionExpiration     = 15 * time.Minute
	messageExpiration     = 1 * time.Hour
	userSessionExpiration = 10 * time.Minute
)

type RedisStore struct {
	client *redis.Client
	store  store.Store
}

func New(cfg *config.Config, underlying store.Store) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		store:  underlying,
	}, nil
}

func (s *RedisStore) Close() {
	s.client.Close()
	s.store.Close()
}

// Cache operations
func (s *RedisStore) setCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, expiration).Err()
}

func (s *RedisStore) getFromCache(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *RedisStore) invalidateCache(ctx context.Context, keys ...string) {
	if len(keys) > 0 {
		s.client.Del(ctx, keys...)
	}
}

// User operations
func (s *RedisStore) CreateUser(ctx context.Context, user *models.User) error {
	if err := s.store.CreateUser(ctx, user); err != nil {
		return err
	}
	return s.cacheUser(ctx, user)
}

func (s *RedisStore) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	return s.store.GetUsersByIDs(ctx, ids)
}

func (s *RedisStore) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.store.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.cacheUser(ctx, user)
}

func (s *RedisStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.store.DeleteUser(ctx, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, fmt.Sprintf(userKey, id))
	return nil
}

// Session operations
func (s *RedisStore) CreateSession(ctx context.Context, session *models.Session) error {
	if err := s.store.CreateSession(ctx, session); err != nil {
		return err
	}
	return s.cacheSession(ctx, session)
}

func (s *RedisStore) GetSessionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Session, error) {
	return s.store.GetSessionsByIDs(ctx, ids)
}

func (s *RedisStore) UpdateSession(ctx context.Context, session *models.Session) error {
	if err := s.store.UpdateSession(ctx, session); err != nil {
		return err
	}
	return s.cacheSession(ctx, session)
}

func (s *RedisStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	if err := s.store.DeleteSession(ctx, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, fmt.Sprintf(sessionKey, id))
	return nil
}

// Message operations
func (s *RedisStore) CreateMessage(ctx context.Context, message *models.Message) error {
	if err := s.store.CreateMessage(ctx, message); err != nil {
		return err
	}
	return s.cacheMessage(ctx, message)
}

func (s *RedisStore) GetMessagesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Message, error) {
	return s.store.GetMessagesByIDs(ctx, ids)
}

func (s *RedisStore) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	if err := s.store.DeleteMessage(ctx, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, fmt.Sprintf(messageKey, id))
	return nil
}

// UserSession operations
func (s *RedisStore) AddUserToSession(ctx context.Context, userID, sessionID uuid.UUID, role string) error {
	if err := s.store.AddUserToSession(ctx, userID, sessionID, role); err != nil {
		return err
	}
	s.invalidateCache(ctx,
		fmt.Sprintf(sessionKey, sessionID),
		fmt.Sprintf(userSessionsKey, userID),
	)
	return nil
}

func (s *RedisStore) RemoveUserFromSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if err := s.store.RemoveUserFromSession(ctx, userID, sessionID); err != nil {
		return err
	}
	s.invalidateCache(ctx,
		fmt.Sprintf(sessionKey, sessionID),
		fmt.Sprintf(userSessionsKey, userID),
	)
	return nil
}

func (s *RedisStore) GetSessionIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.store.GetSessionIDsByUserID(ctx, userID)
}

func (s *RedisStore) GetUserIDsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]uuid.UUID, error) {
	return s.store.GetUserIDsBySessionID(ctx, sessionID)
}

// Pass-through methods
func (s *RedisStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.store.GetUserByUsername(ctx, username)
}

func (s *RedisStore) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return s.store.CheckUsernameExists(ctx, username)
}

func (s *RedisStore) CheckNicknameExists(ctx context.Context, nickname string) (bool, error) {
	return s.store.CheckNicknameExists(ctx, nickname)
}

func (s *RedisStore) GetMessageIDsBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, before time.Time) ([]uuid.UUID, error) {
	return s.store.GetMessageIDsBySessionID(ctx, sessionID, limit, before)
}

func (s *RedisStore) BeginTx(ctx context.Context) (store.Transaction, error) {
	return s.store.BeginTx(ctx)
}

// Cache helpers
func (s *RedisStore) cacheUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf(userKey, user.ID)
	cachedUser := CachedUser{
		ID:        user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
	return s.setCache(ctx, key, cachedUser, userExpiration)
}

func (s *RedisStore) cacheSession(ctx context.Context, session *models.Session) error {
	key := fmt.Sprintf(sessionKey, session.ID)
	// Get current members if they exist
	var cached CachedSession
	_ = s.getFromCache(ctx, key, &cached)

	cached.Session = session
	return s.setCache(ctx, key, cached, sessionExpiration)
}

func (s *RedisStore) cacheMessage(ctx context.Context, message *models.Message) error {
	key := fmt.Sprintf(messageKey, message.ID)
	return s.setCache(ctx, key, message, messageExpiration)
}
