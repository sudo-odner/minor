export interface User {
    id: string;
    username: string;
    email: string;
    avatar_url?: string;
    bio?: string;
}

export interface RelationshipPreview {
    user_id: string;
    username: string;
    avatar_url?: string;
    status: 'friends' | 'request_sent' | 'request_received' | 'blocked';
}
