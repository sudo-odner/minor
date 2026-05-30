CREATE TABLE IF NOT EXISTS relationships (
    user_id   UUID REFERENCES users(id) ON DELETE CASCADE,
    target_id UUID REFERENCES users(id) ON DELETE CASCADE,

   -- Статус отношений:
    -- 1 - friends (друзья)
    -- 2 - request_sent (заявка отправлена)
    -- 3 - request_received (ожидание подтверждения)
    -- 4 - blocked (пользователь заблокирован)
    status     SMALLINT NOT NULL DEFAULT 2,

    create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, target_id)
);

CREATE INDEX idx_relationships_friend_status ON relationships(target_id, status);
