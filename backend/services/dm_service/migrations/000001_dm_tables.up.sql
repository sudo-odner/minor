-- DM Channel
create table channels (
    id uuid primary key,
    type dm_channel_type not null,
    name varchar(255),                                      -- Only for dm_group
    updated_at timestamp not null default current_timestamp,
    created_at timestamp not null default current_timestamp
);

-- DM members
create table members_channel (
    user_id uuid not null,
    channel_id uuid not null references channels(id) on delete cascade,
    created_at timestamp not null default current_timestamp,

    primary key (user_id, channel_id)
);

create index idx_members_user_id on members_channel(user_id);
create index idx_members_channel_id on members_channel(channel_id);
