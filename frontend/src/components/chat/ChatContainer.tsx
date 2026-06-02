import React, { useEffect, useState, useRef } from 'react';
import { getMessages, Message, sendMessage } from '../../api/messages';
import { useSocket } from '../../context/SocketContext';
import MessageInput from './MessageInput';

interface ChatContainerProps {
  channelId: string;
  channelName: string;
}

const ChatContainer: React.FC<ChatContainerProps> = ({ channelId, channelName }) => {
  const [messages, setServers] = useState<Message[]>([]);
  const { socket } = useSocket();
  const scrollRef = useRef<HTMLDivElement>(null);

  // 1. Загрузка истории при смене канала
  useEffect(() => {
    const loadHistory = async () => {
      try {
        const history = await getMessages(channelId);
        // Cassandra отдает от новых к старым, для UI инвертируем
        setServers(history.reverse());
        scrollToBottom();
      } catch (err) {
        console.error("Failed to load history", err);
      }
    };
    loadHistory();
  }, [channelId]);

  // 2. Прослушивание WebSocket для новых сообщений
  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      const payload = JSON.parse(event.data);
      
      // Проверяем, что это событие создания сообщения и оно для текущего канала
      if (payload.t === 'MESSAGE_CREATE' && payload.d.channel_id === channelId) {
        setServers((prev) => [...prev, payload.d]);
        scrollToBottom();
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
    <div className="flex flex-col h-full bg-[#313338]">
      {/* Шапка канала */}
      <div className="h-12 flex items-center px-4 shadow-sm border-b border-[#1e1f22] shrink-0">
        <span className="text-gray-400 text-2xl mr-2">#</span>
        <span className="font-bold text-white">{channelName}</span>
      </div>

      {/* Список сообщений */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
        {messages.map((msg) => (
          <div key={msg.id} className="flex items-start space-x-3 hover:bg-[#2e3035] -mx-4 px-4 py-1 group">
            <div className="w-10 h-10 bg-gray-600 rounded-full shrink-0 flex items-center justify-center font-bold">
              {msg.username?.[0].toUpperCase() || 'U'}
            </div>
            <div>
              <div className="flex items-baseline space-x-2">
                <span className="font-bold text-white hover:underline cursor-pointer">
                  {msg.username || 'Аноним'}
                </span>
                <span className="text-xs text-gray-400">
                  {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              <p className="text-gray-300 whitespace-pre-wrap">{msg.content}</p>
            </div>
          </div>
        ))}
        <div ref={scrollRef} />
      </div>

      {/* Поле ввода */}
      <MessageInput onSend={handleSend} placeholder={`Написать в #${channelName}`} />
    </div>
  );
};

export default ChatContainer;