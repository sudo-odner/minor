import React, { useEffect, useState, useRef } from 'react';
import { getMessages, Message, sendMessage } from '../../api/messages';
import { useSocket } from '../../context/SocketContext';
import { useAuth } from '../../context/AuthContext';
import MessageInput from './MessageInput';
import { api } from '../../api/axios';

interface ChatContainerProps {
  channelId: string;
  channelName: string;
}

const ChatContainer: React.FC<ChatContainerProps> = ({ channelId, channelName }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [userCache, setUserCache] = useState<Record<string, { username: string; avatarUrl?: string }>>({});
  const { socket } = useSocket();
  const { user } = useAuth();
  const scrollRef = useRef<HTMLDivElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 1. Загрузка истории при смене канала
  useEffect(() => {
    const loadHistory = async () => {
      try {
        const history = await getMessages(channelId);
        // Cassandra отдает от новых к старым, для UI инвертируем
        setMessages([...history].reverse());
      } catch (err) {
        console.error("Failed to load history", err);
      }
    };
    loadHistory();
  }, [channelId]);

  // 2. Авто-скролл вниз при новых сообщениях
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // 3. Загрузка недостающих профилей пользователей
  useEffect(() => {
    const fetchMissingProfiles = async () => {
      const missingIds = Array.from(
        new Set(
          messages
            .map((m) => m.author_id)
            .filter((id) => id && !userCache[id])
        )
      );
      
      if (missingIds.length === 0) return;

      const newProfiles: Record<string, { username: string; avatarUrl?: string }> = {};
      let updated = false;

      await Promise.all(
        missingIds.map(async (id) => {
          try {
            const res = await api.get(`/users/${id}`);
            if (res.data && res.data.username) {
              newProfiles[id] = {
                username: res.data.username,
                avatarUrl: res.data.avatar_url,
              };
              updated = true;
            }
          } catch (err) {
            console.error(`Failed to fetch profile for user ${id}:`, err);
          }
        })
      );

      if (updated) {
        setUserCache(prev => ({ ...prev, ...newProfiles }));
      }
    };

    fetchMissingProfiles();
  }, [messages, userCache]);

  // 4. Прослушивание WebSocket для новых сообщений
  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data);
        
        // В нашем протоколе события приходят с op: 1
        if (payload.op === 1 && payload.t === 'MESSAGE_CREATE') {
          const newMsg = payload.d;
          if (newMsg.channel_id === channelId) {
            setMessages((prev) => {
              // Предотвращаем дубликаты
              if (prev.some(m => m.message_id === newMsg.message_id)) return prev;
              return [...prev, newMsg];
            });
          }
        }
      } catch (err) {
        console.error("Failed to parse WS payload:", err);
      }
    };

    socket.addEventListener('message', handleMessage);
    return () => socket.removeEventListener('message', handleMessage);
  }, [socket, channelId]);

  const handleSend = async (content: string) => {
    try {
      await sendMessage(channelId, content);
      // Мы не добавляем сообщение в стейт здесь вручную, 
      // оно прилетит к нам через WebSocket (как в Discord),
      // что гарантирует синхронизацию всех клиентов.
    } catch (err) {
      alert("Ошибка отправки");
    }
  };

  return (
    <div className="flex flex-col h-full bg-white">
      {/* Шапка канала */}
      <div className="h-12 flex items-center px-4 shadow-sm border-b border-brand-blue-light shrink-0 bg-white">
        <span className="text-brand-blue text-2xl mr-2 opacity-50 font-light">#</span>
        <span className="font-bold text-gray-800">{channelName}</span>
      </div>

      {/* Список сообщений */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
        {messages.map((msg) => {
          const isMe = msg.author_id === user?.id;
          const authorName = msg.username || userCache[msg.author_id]?.username || 'Аноним';
          
          return (
            <div 
              key={msg.message_id} 
              className={`flex items-start space-x-3 hover:bg-brand-blue-light/20 -mx-4 px-4 py-1 group transition-colors ${
                isMe ? 'flex-row-reverse space-x-reverse' : ''
              }`}
            >
              <div className="w-10 h-10 bg-brand-blue/10 text-brand-blue rounded-full shrink-0 flex items-center justify-center font-bold border border-brand-blue-light">
                  {authorName.charAt(0).toUpperCase()}
              </div>
              <div className={isMe ? 'text-right' : ''}>
                <div className={`flex items-baseline space-x-2 ${isMe ? 'flex-row-reverse space-x-reverse' : ''}`}>
                  <span className="font-bold text-gray-800 hover:underline cursor-pointer">
                    {authorName}
                  </span>
                  <span className="text-[10px] text-gray-400 font-medium uppercase">
                    {new Date(msg.create_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <p className="text-gray-600 whitespace-pre-wrap text-sm leading-relaxed">{msg.content}</p>
              </div>
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      {/* Поле ввода */}
      <MessageInput onSend={handleSend} placeholder={`Написать в #${channelName}`} />
    </div>
  );
};

export default ChatContainer;
