import React, { useEffect, useState, useRef } from 'react';
import { getMessages, Message, sendMessage } from '../../api/messages';
import { useSocket } from '../../context/SocketContext';
import { useAuth } from '../../context/AuthContext';
import MessageInput from './MessageInput';

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

  // 1. Загрузка истории при смене канала
  useEffect(() => {
    const loadHistory = async () => {
      try {
        const history = await getMessages(channelId);
        // Cassandra отдает от новых к старым, для UI инвертируем
        setMessages(history.reverse());
        scrollToBottom();
      } catch (err) {
        console.error("Failed to load history", err);
      }
    };
    loadHistory();
  }, [channelId]);

  // 2. Загрузка недостающих профилей пользователей для отображения ников
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

      const newProfiles = { ...userCache };
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
        setUserCache(newProfiles);
      }
    };

    fetchMissingProfiles();
  }, [messages, userCache]);

  // 3. Прослушивание WebSocket для новых сообщений
  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data);
        
        // Проверяем, что это событие создания сообщения и оно для текущего канала
        if (payload.t === 'MESSAGE_CREATE' && payload.d.channel_id === channelId) {
          setMessages((prev) => [...prev, payload.d]);
          scrollToBottom();
        }
      } catch (err) {
        console.error("Failed to parse WS payload:", err);
      }
    };

    socket.addEventListener('message', handleMessage);
    return () => socket.removeEventListener('message', handleMessage);
  }, [socket, channelId]);

  const scrollToBottom = () => {
    setTimeout(() => {
      scrollRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, 100);
  };

  const handleSend = async (content: string) => {
    try {
      await sendMessage(channelId, content);
      // Мы не добавляем сообщение в стейт здесь, 
      // оно прилетит к нам через WebSocket (как в Discord)
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
          
          return (
            <div 
              key={msg.message_id} 
              className={`flex items-start space-x-3 hover:bg-brand-blue-light/20 -mx-4 px-4 py-1 group transition-colors ${
                isMe ? 'flex-row-reverse space-x-reverse' : ''
              }`}
            >
              <div className="w-10 h-10 bg-brand-blue/10 text-brand-blue rounded-full shrink-0 flex items-center justify-center font-bold border border-brand-blue-light">
                  {(msg.username?.charAt(0) || 'U').toUpperCase()}
              </div>
              <div className={isMe ? 'text-right' : ''}>
                <div className={`flex items-baseline space-x-2 ${isMe ? 'flex-row-reverse space-x-reverse' : ''}`}>
                  <span className="font-bold text-gray-800 hover:underline cursor-pointer">
                    {msg.username || 'Аноним'}
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
        <div ref={scrollRef} />
      </div>

      {/* Поле ввода */}
      <MessageInput onSend={handleSend} placeholder={`Написать в #${channelName}`} />
    </div>
  );
};

export default ChatContainer;