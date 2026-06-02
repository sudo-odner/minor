export interface Server {
    id: string;
    name: string;
    owner_id: string;
    avatar_url?: string;
}

export interface Channel {
    id: string;
    server_id: string;
    name: string;
    type: number;
    parent_id?: string;
    position: number;
}

export interface Message {
    message_id: string;
    channel_id: string;
    author_id: string;
    content: string;
    reply_to?: string;
    create_at: string;
}
