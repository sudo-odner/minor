CREATE TABLE servers (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner_id UUID NOT NULL,
    icon_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    permissions BIGINT NOT NULL DEFAULT 0,
    color INTEGER DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE channels (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name VARCHAR(32) NOT NULL,
    type SMALLINT NOT NULL DEFAULT 0, -- 0-текст, 1-голос
    parent_id UUID REFERENCES channels(id), 
    position INTEGER DEFAULT 0
);

CREATE TABLE server_members (
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    nickname VARCHAR(32),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (server_id, user_id)
);

CREATE TABLE member_roles (
    server_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (server_id, user_id, role_id),
    FOREIGN KEY (server_id, user_id) REFERENCES server_members(server_id, user_id) ON DELETE CASCADE
);