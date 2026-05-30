CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    username   VARCHAR(50) NOT NULL UNIQUE,
    avatar_url VARCHAR(512),
    bio        TEXT NOT NULL,
    create_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    update_at  TIMESTAMP NOT NULL DEFAULT NOW()

    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT check_username_len CHECK (char_length(username) >= 2)
);
