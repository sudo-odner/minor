import { api } from './axios';

export interface Message {
  id: string;         // UUID v7
  channel_id: string;
  author_id: string;
  content: string;
  created_at: string;
  username?: string;  // Обогащенное поле
  avatar_url?: string;
}

// Загрузка истории (курсорная пагинация)
export const getMessages = async (channelId: string, before?: string): Promise<Message[]> => {
  const url = before 
    ? `/messages/${channelId}?before=${before}&limit=50` 
    : `/messages/${channelId}?limit=50`;
  const response = await api.get(url);
  return Array.isArray(response.data) ? response.data : [];
};

// Отправка сообщения
export const sendMessage = async (channelId: string, content: string): Promise<Message> => {
  const response = await api.post(`/messages/${channelId}`, { content });
  return response.data;
};