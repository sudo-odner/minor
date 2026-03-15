CREATE TYPE friend_status AS ENUM ('pending', 'accepted', 'deny', 'blocked');

CREATE TABLE IF NOT EXISTS friends (
    user_id   UUID REFERENCES users(id),
    friend_id UUID REFERENCES users(id),
    status    friend_status NOT NULL,
    create_at TIMESTAMP NOT NULL DEFAULT Now(),
    PRIMARY KEY(user_id, friend_id)
);
