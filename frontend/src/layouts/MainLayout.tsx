import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useSocket } from '../context/SocketContext';
import { api } from '../api/axios';
import CreateChannelModal from '../components/modals/CreateChannelModal';
import CreateServerModal from '../components/modals/CreateServerModal'; // Импортируем модалку сервера

const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { logout, user, accessToken } = useAuth();
  const { isConnected } = useSocket();
  
  const [servers, setServers] = useState<any[]>([]);
  const [activeServerId, setActiveServerId] = useState<string | null>(null);
  const [channels, setChannels] = useState<any[]>([]);
  
  // Состояния открытия модальных окон
  const [isChannelModalOpen, setIsChannelModalOpen] = useState(false);
  const [isServerModalOpen, setIsServerModalOpen] = useState(false);

  // 1. Загрузка серверов пользователя при старте
  const fetchServers = async () => {
    if (!accessToken) return;
    try {
      const res = await api.get('/servers/@me');
      const serverList = Array.isArray(res.data) ? res.data : (res.data?.servers || []);
      setServers(serverList);
      
      if (serverList.length > 0 && !activeServerId) {
        setActiveServerId(serverList[0].id);
      }
    } catch (err) {
      console.error("Failed to fetch servers", err);
    }
  };

  useEffect(() => {
    fetchServers();
  }, [accessToken]);

  // 2. Загрузка каналов при смене активного сервера
  useEffect(() => {
    if (!accessToken || !activeServerId) return;

    const fetchChannels = async () => {
      try {
        const res = await api.get(`/servers/${activeServerId}/channels`);
        const channelList = Array.isArray(res.data) ? res.data : (res.data?.channels || []);
        setChannels(channelList);
      } catch (err) {
        console.error("Failed to fetch channels", err);
        setChannels([]);
      }
    };
    fetchChannels();
  }, [accessToken, activeServerId]);

  // Callback для добавления созданного канала
  const handleChannelCreated = (newChannel: any) => {
    setChannels((prev) => [...prev, newChannel]);
  };

  // Callback для добавления созданного сервера
  const handleServerCreated = (newServer: any) => {
    setServers((prev) => [...prev, newServer]);
    setActiveServerId(newServer.id); // Переключаемся на новый сервер
  };

  return (
    <div className="flex h-screen w-full bg-[#313338] text-white overflow-hidden font-sans">
      
      {/* 1. Панель серверов */}
      <div className="w-[72px] bg-[#1e1f22] flex flex-col items-center py-3 space-y-2 overflow-y-auto no-scrollbar">
        <div 
          onClick={() => setActiveServerId(null)}
          className={`w-12 h-12 rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl transition-all ${
            !activeServerId ? 'bg-[#5865f2] rounded-xl' : 'bg-[#313338] hover:bg-[#5865f2]'
          }`}
        >
          <span className="font-bold text-xl">M</span>
        </div>
        <div className="w-8 h-[2px] bg-[#35363c] rounded-full mx-auto my-1" />
        
        {servers.map(server => (
          <div 
            key={server.id} 
            onClick={() => setActiveServerId(server.id)}
            className={`w-12 h-12 flex items-center justify-center cursor-pointer transition-all group ${
              activeServerId === server.id 
                ? 'bg-[#5865f2] rounded-xl' 
                : 'bg-[#313338] rounded-full hover:rounded-xl hover:bg-[#5865f2]'
            }`}
          >
            <span className="text-sm font-semibold truncate px-1 uppercase">{server.name.substring(0, 3)}</span>
          </div>
        ))}

        {/* Кнопка "Создать сервер" */}
        <div 
          onClick={() => setIsServerModalOpen(true)}
          className="w-12 h-12 bg-[#313338] rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl hover:bg-[#23a559] transition-all text-[#23a559] hover:text-white"
          title="Создать сервер"
        >
          <span className="text-2xl">+</span>
        </div>
      </div>

      {/* 2. Панель каналов */}
      <div className="w-60 bg-[#2b2d31] flex flex-col">
        <div className="h-12 shadow-sm flex items-center px-4 font-bold border-b border-[#1e1f22] justify-between">
          <span className="truncate">
            {activeServerId 
              ? servers.find(s => s.id === activeServerId)?.name 
              : "Личные сообщения"
            }
          </span>
        </div>
        
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {activeServerId ? (
            <>
              <div className="flex items-center justify-between text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">
                <span>Каналы</span>
                <button 
                  onClick={() => setIsChannelModalOpen(true)}
                  className="hover:text-white text-lg transition-colors focus:outline-none"
                  title="Создать канал"
                >
                  +
                </button>
              </div>

              {channels.map(channel => (
                <div key={channel.id} className="px-2 py-1.5 rounded hover:bg-[#35373c] cursor-pointer text-gray-400 hover:text-gray-100 flex items-center justify-between group">
                  <div className="flex items-center truncate">
                    <span className="mr-2 text-xl text-gray-500 font-bold">{channel.type === 0 ? '#' : '🔊'}</span> 
                    <span className="truncate text-sm">{channel.name}</span>
                  </div>
                </div>
              ))}

              {channels.length === 0 && (
                <div className="px-2 py-4 text-xs text-gray-500 italic">Каналов пока нет...</div>
              )}
            </>
          ) : (
            <div className="px-2 py-4 text-xs text-gray-500 italic">Тут будут личные переписки</div>
          )}
        </div>

        {/* Профиль пользователя */}
        <div className="bg-[#232428] p-2 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="relative">
              <div className="w-8 h-8 bg-gray-600 rounded-full" />
              <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#232428] ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            </div>
            <div className="text-xs overflow-hidden w-24">
              <p className="font-bold truncate">{user?.username || 'User'}</p>
              <p className="text-gray-400 truncate">Online</p>
            </div>
          </div>
          <button onClick={logout} className="text-gray-400 hover:text-red-400 p-1">
             🚪
          </button>
        </div>
      </div>

      {/* 3. Основная область чата */}
      <main className="flex-1 flex flex-col bg-[#313338]">
        {children}
      </main>

      {/* Модальное окно создания канала */}
      {activeServerId && (
        <CreateChannelModal 
          isOpen={isChannelModalOpen}
          onClose={() => setIsChannelModalOpen(false)}
          serverId={activeServerId}
          onChannelCreated={handleChannelCreated}
        />
      )}

      {/* Модальное окно создания сервера */}
      <CreateServerModal 
        isOpen={isServerModalOpen}
        onClose={() => setIsServerModalOpen(false)}
        onServerCreated={handleServerCreated}
      />
    </div>
  );
};

export default MainLayout;
