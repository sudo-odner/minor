-- Таблица сервера. Хранит в себе metadata о серверe
CREATE TABLE servers (
    id         UUID PRIMARY KEY,
    name       VARCAHR(100) NOT NULL,
    owner_id   UUID NOT NULL, -- Пользователь который владеет каналом
    avatar_url VARCHAR(512),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);

-- Таблица участриков. Связывает пользователей с серверами(+локальное имя на сервере)
CREATE TABLE members (
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    user_id   UUID NOT NULL, 
    nickname  VARCHAR(100),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (server_id, user_id)
);

-- Таблица ролей. Указывает роли сервера и какие права доступа имеет
CREATE TABLE roles (
    id          UUID PRIMARY KEY,
    server_id   UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name        VARCHAR(50) NOT NULL,
    permissions BIGINT NOT NULL DEFAULT 0,
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Связующая таблица участников и ролей на сервере
CREATE TABLE members_roles (
    server_id UUID NOT NULL,
    user_id   UUID NOT NULL,
    role_id   UUID NOT NULL,

    PRIMARY KEY (server_id, user_id, role_id),
    FOREIGN KEY (server_id, user_id) REFERENCES members(server_id, user_id) ON DELETE CASCADE
);

-- Каланы и категори. Харнит каналы, где могут групироватся по категории
CREATE TABLE channels (
    id         UUID PRIMARY KEY,
    server_id  UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name       VARCHAR(64) NOT NULL,
    type       SMALLINT NOT NULL DEFAULT 0, -- 0-category, 1-text, 2-voice

    parent_id  UUID REFERENCES channels(id) ON DELETE SET NULL,

    position   INTEGER NOT NULL DEFAULT 0,
    created_at TIEMSTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Для того чтобы быстро сортировать по server_id
CREATE INDEX idx_channels_server ON channels(server_id);

CREATE TYPE override_target_type AS ENUM ('role', 'user');
-- Переопределение, когда нам нужно чтобы определенные каналы было доступны только некоторым или наоборот.
CREATE TABLE channel_permission_overrides (
    channel_id     UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,

    target_type    override_target_type NOT NULL, -- 'role' or 'user'
    target_id UUID NOT NULL, --'role_id' or 'user_id'

    allow          BIGINT NOT NULL DEFAULT 0,
    deny           BIGINT NOT NULL DEFAULT 0,
    
    PRIMARY KEY (channel_id, target_id)
);
