CREATE TABLE credentials (
    id BIGINT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);