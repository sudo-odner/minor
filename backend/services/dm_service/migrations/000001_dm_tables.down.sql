drop index if exists idx_members_user_id;
drop index if exists idx_members_channel_id;

drop table if exists members_channel;
drop table if exists channels;

drop type if exists dm_channel_type;
