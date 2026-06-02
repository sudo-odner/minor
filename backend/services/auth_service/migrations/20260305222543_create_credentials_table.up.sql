CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true, 
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);