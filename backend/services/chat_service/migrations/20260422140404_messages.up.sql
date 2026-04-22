CREATE TABLE IF NOT EXISTS messages (
    channel_id uuid,
    message_id uuid,
    author_id uuid,
    content text,
    reply_to uuid,
    create_at timestamp,
    PRIMARY KEY (channel_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
