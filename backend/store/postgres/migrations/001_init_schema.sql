-- Create initial schema for the chat application
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    username    TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    nickname    TEXT NOT NULL UNIQUE,
    avatar_url  TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE sessions (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    creator_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE user_sessions (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id  UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role        TEXT NOT NULL,
    joined_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (user_id, session_id)
);

CREATE TABLE messages (
    id          UUID PRIMARY KEY,
    type        TEXT NOT NULL,
    content     TEXT NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id  UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    timestamp   TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX messages_session_id_timestamp_idx ON messages(session_id, timestamp DESC);
CREATE INDEX user_sessions_user_id_idx ON user_sessions(user_id);
CREATE INDEX user_sessions_session_id_idx ON user_sessions(session_id);

-- Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users; 