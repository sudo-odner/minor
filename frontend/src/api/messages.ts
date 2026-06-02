import { api } from './axios';

export interface Message {
  message_id: string; // В Go: MessageID
  channel_id: string; // В Go: ChannelID
  author_id: string;  // В Go: AuthorID
  content: string;
  create_at: string;  // ИСПРАВЛЕНО: убрали 'd', теперь совпадает с Go
  username?: string;  // Это поле мы добавим на бэкенде
}

// Загрузка истории (курсорная пагинация)
// src/api/messages.ts

export const getMessages = async (channelId: string, beforeId?: string): Promise<Message[]> => {
  // 1. Используем messages (множественное число)
  // 2. Используем ключ before_id (как на бэкенде)
  const url = beforeId 
    ? `/messages/${channelId}?before_id=${beforeId}&limit=50` 
    : `/messages/${channelId}?limit=50`;
    
  const response = await api.get(url);
  
  // 3. Возвращаем именно поле messages из обертки ResGetMessages
  return response.data.messages || [];
};

export const sendMessage = async (channelId: string, content: string): Promise<Message> => {
  // Добавляем слэш в конце
  const response = await api.post(`/messages/${channelId}/`, { content });
  return response.data;
};