import React, { useEffect, useState, useRef } from 'react';
import { getMessages, Message, sendMessage } from '../../api/messages';
import { useSocket } from '../../context/SocketContext';
import { api } from '../../api/axios';
import MessageInput from './MessageInput';

interface ChatContainerProps {
  channelId: string;
  channelName: string;
}

const ChatContainer: React.FC<ChatContainerProps> = ({ channelId, channelName }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [userCache, setUserCache] = useState<Record<string, { username: string; avatarUrl?: string }>>({});
  const { socket } = useSocket();
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
    <div className="flex flex-col h-full bg-white dark:bg-[#313338] transition-colors duration-200">
      {/* Шапка канала */}
      <div className="h-12 flex items-center px-4 shadow-sm border-b border-[#e3e5e8] dark:border-[#1e1f22] shrink-0 transition-colors">
        <span className="text-[#4f5660] dark:text-gray-400 text-2xl mr-2">#</span>
        <span className="font-bold text-[#060607] dark:text-white">{channelName}</span>
      </div>

      {/* Список сообщений */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
        {messages.map((msg) => {
          const authorName = userCache[msg.author_id]?.username || msg.username || 'Аноним';
          const avatarLetter = authorName.charAt(0).toUpperCase() || 'U';
          return (
            <div key={msg.message_id} className="flex items-start space-x-3 hover:bg-[#f2f3f5] dark:hover:bg-[#2e3035] -mx-4 px-4 py-1 group transition-colors">
              <div className="w-10 h-10 bg-[#5865f2] rounded-full shrink-0 flex items-center justify-center font-bold text-white">
                {avatarLetter}
              </div>
              <div>
                <div className="flex items-baseline space-x-2">
                  <span className="font-bold text-[#060607] dark:text-white hover:underline cursor-pointer">
                    {authorName}
                  </span>
                  <span className="text-xs text-[#4f5660] dark:text-gray-400">
                    {new Date(msg.create_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <p className="text-[#2e3338] dark:text-gray-300 whitespace-pre-wrap">{msg.content}</p>
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