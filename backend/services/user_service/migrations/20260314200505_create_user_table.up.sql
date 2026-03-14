CREATE TABLE IF NOT EXISTS users (
    id        UUID PRIMARY KEY,
    username  VARCHAR(50) NOT NULL,
    bio       TEXT NOT NULL,
    create_at TIMESTAMP NOT NULL DEFAULT NOW(),
    update_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS index_users_username ON users(username);
