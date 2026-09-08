-- Enum type channel
create type dm_channel_type as enum('dm', 'dm_group');

-- DM Channel
create table channels (
    id uuid primary key,
    type dm_channel_type not null,
    updated_at timestamp not null default current_timestamp,
    created_at timestamp not null default current_timestamp
);

-- DM members
create table members_channel (
    id uuid primary key,
    user_id uuid not null,
    channel_id uuid not null,
    name varchar(255),                                      -- For DM user can create alias on chat, for DM_GROUP is name chat
    updated_at timestamp not null default current_timestamp,
    created_at timestamp not null default current_timestamp
);
