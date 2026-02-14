CREATE DATABASE secure_auth;
\c secure_auth;

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

CREATE INDEX idx_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_token ON refresh_tokens(token);