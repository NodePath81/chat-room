-- Add additional indexes for better query performance
CREATE INDEX users_username_idx ON users(username);
CREATE INDEX users_nickname_idx ON users(nickname);
CREATE INDEX messages_user_id_idx ON messages(user_id);
CREATE INDEX messages_created_at_idx ON messages(created_at DESC);

-- Add partial index for active sessions (non-deleted)
CREATE INDEX active_sessions_idx ON sessions(created_at DESC, updated_at DESC);

-- Add btree_gin extension for faster text search
CREATE EXTENSION IF NOT EXISTS btree_gin;

-- Add GIN indexes for text search
CREATE INDEX messages_content_gin_idx ON messages USING gin(to_tsvector('english', content));
CREATE INDEX sessions_name_gin_idx ON sessions USING gin(to_tsvector('english', name));

-- Down
DROP INDEX IF EXISTS messages_content_gin_idx;
DROP INDEX IF EXISTS sessions_name_gin_idx;
DROP EXTENSION IF EXISTS btree_gin;
DROP INDEX IF EXISTS active_sessions_idx;
DROP INDEX IF EXISTS messages_created_at_idx;
DROP INDEX IF EXISTS messages_user_id_idx;
DROP INDEX IF EXISTS users_nickname_idx;
DROP INDEX IF EXISTS users_username_idx; 