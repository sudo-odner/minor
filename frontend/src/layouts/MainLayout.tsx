import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useSocket } from '../context/SocketContext';
import { api } from '../api/axios';

// Импорт компонентов
import CreateServerModal from '../components/modals/CreateServerModal';
import JoinServerModal from '../components/modals/JoinServerModal';
import CreateChannelModal from '../components/modals/CreateChannelModal';
import MemberList from '../components/server/MemberList';
import ChatContainer from '../components/chat/ChatContainer';

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout = ({ children }: MainLayoutProps) => {
  const { logout, user, accessToken } = useAuth();
  const { isConnected } = useSocket();

  // Списки данных
  const [servers, setServers] = useState<any[]>([]);
  const [channels, setChannels] = useState<any[]>([]);
  
  // Состояния выбора
  const [activeServer, setActiveServer] = useState<any | null>(null);
  const [activeChannel, setActiveChannel] = useState<any | null>(null);

  // Состояния модалок
  const [isCreateServerOpen, setIsCreateServerOpen] = useState(false);
  const [isJoinServerOpen, setIsJoinServerOpen] = useState(false);
  const [isCreateChannelOpen, setIsCreateChannelOpen] = useState(false);

  // 1. Загрузка серверов пользователя
  useEffect(() => {
    if (!accessToken) return;

    const fetchServers = async () => {
      try {
        const res = await api.get('/servers/@me');
        const data = Array.isArray(res.data) ? res.data : [];
        setServers(data);
        
        // Автоматически выбираем первый сервер, если ничего не выбрано
        if (data.length > 0 && !activeServer) {
          setActiveServer(data[0]);
        }
      } catch (err) {
        console.error('Ошибка загрузки серверов:', err);
      }
    };
    fetchServers();
  }, [accessToken]);

  // 2. Загрузка каналов при смене сервера
  useEffect(() => {
    if (!accessToken || !activeServer) {
      setChannels([]);
      setActiveChannel(null);
      return;
    }

    const fetchChannels = async () => {
      try {
        const res = await api.get(`/servers/${activeServer.id}/channels`);
        const data = Array.isArray(res.data) ? res.data : [];
        setChannels(data);
        
        // Выбираем первый канал по умолчанию
        if (data.length > 0) {
          setActiveChannel(data[0]);
        }
      } catch (err) {
        console.error('Ошибка загрузки каналов:', err);
        setChannels([]);
      }
    };
    fetchChannels();
  }, [accessToken, activeServer]);

  // Колбэки для обновления UI
  const handleServerCreated = (newServer: any) => {
    setServers(prev => [...prev, newServer]);
    setActiveServer(newServer);
  };

  const handleChannelCreated = (newChannel: any) => {
    setChannels(prev => [...prev, newChannel]);
    setActiveChannel(newChannel);
  };

  return (
    <div className="flex h-screen w-full bg-[#313338] text-white overflow-hidden font-sans">
      
      {/* ПАНЕЛЬ 1: СПИСОК СЕРВЕРОВ (Самая левая) */}
      <nav className="w-[72px] bg-[#1e1f22] flex flex-col items-center py-3 space-y-2 overflow-y-auto no-scrollbar">
        {/* Кнопка "Личные сообщения" */}
        <div 
          onClick={() => setActiveServer(null)}
          className={`w-12 h-12 rounded-full flex items-center justify-center cursor-pointer transition-all duration-200 ${
            !activeServer ? 'bg-[#5865f2] rounded-xl' : 'bg-[#313338] hover:bg-[#5865f2] hover:rounded-xl'
          }`}
        >
          <span className="text-xl font-bold">M</span>
        </div>

        <div className="w-8 h-[2px] bg-[#35363c] rounded-full mx-auto my-1" />

        {/* Список иконок серверов */}
        {servers.map((server) => (
          <div
            key={server.id}
            onClick={() => setActiveServer(server)}
            className={`relative flex items-center group mb-2`}
          >
            {/* Полоска индикатор слева */}
            <div className={`absolute -left-3 w-1 bg-white rounded-r-full transition-all duration-200 ${
              activeServer?.id === server.id ? 'h-10' : 'h-2 scale-0 group-hover:scale-100 group-hover:h-5'
            }`} />
            
            <div className={`w-12 h-12 flex items-center justify-center cursor-pointer transition-all duration-200 shadow-lg ${
              activeServer?.id === server.id ? 'bg-[#5865f2] rounded-xl' : 'bg-[#313338] rounded-full hover:rounded-xl hover:bg-[#5865f2]'
            }`}>
              <span className="text-sm font-bold uppercase">{(server.name || '??').substring(0, 2)}</span>
            </div>
          </div>
        ))}

        {/* Кнопки действий */}
        <button 
          onClick={() => setIsCreateServerOpen(true)}
          className="w-12 h-12 bg-[#313338] rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl hover:bg-[#23a55a] transition-all text-[#23a55a] hover:text-white"
        >
          <span className="text-2xl font-light">+</span>
        </button>

        <button 
          onClick={() => setIsJoinServerOpen(true)}
          className="w-12 h-12 bg-[#313338] rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl hover:bg-[#5865f2] transition-all text-gray-400 hover:text-white"
        >
          <span className="text-xl">🧭</span>
        </button>
      </nav>

      {/* ПАНЕЛЬ 2: СПИСОК КАНАЛОВ (Средняя) */}
      <aside className="w-60 bg-[#2b2d31] flex flex-col shrink-0">
        <header className="h-12 shadow-md flex items-center px-4 font-bold border-b border-[#1e1f22] z-10">
          <span className="truncate">{activeServer ? activeServer.name : "Личные сообщения"}</span>
        </header>

        <div className="flex-1 overflow-y-auto p-2">
          {activeServer ? (
            <>
              <div className="flex items-center justify-between text-xs font-bold text-gray-400 px-2 mt-4 mb-2 uppercase tracking-wider">
                <span>Текстовые каналы</span>
                <button onClick={() => setIsCreateChannelOpen(true)} className="hover:text-white text-xl leading-none">+</button>
              </div>
              
              <div className="space-y-0.5">
                {channels.map((channel) => (
                  <div
                    key={channel.id}
                    onClick={() => setActiveChannel(channel)}
                    className={`group flex items-center px-2 py-1.5 rounded cursor-pointer transition-colors ${
                      activeChannel?.id === channel.id ? 'bg-[#3f4248] text-white' : 'text-gray-400 hover:bg-[#35373c] hover:text-gray-200'
                    }`}
                  >
                    <span className="text-gray-500 text-xl mr-1.5">#</span>
                    <span className="truncate font-medium">{channel.name}</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <div className="mt-4 px-2 text-sm text-gray-500 text-center italic">
              Выберите сервер для просмотра каналов
            </div>
          )}
        </div>

        {/* ПАНЕЛЬ ПРОФИЛЯ (Футер средней колонки) */}
        <footer className="bg-[#232428] h-[52px] px-2 flex items-center justify-between shrink-0">
          <div className="flex items-center space-x-2 min-w-0">
            <div className="relative shrink-0">
              <div className="w-8 h-8 bg-[#5865f2] rounded-full flex items-center justify-center font-bold text-xs">
                  {(user?.username?.charAt(0) || 'U').toUpperCase()}
              </div>
              <div className={`absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-[3px] border-[#232428] ${
                isConnected ? 'bg-[#23a55a]' : 'bg-[#80848e]'
              }`} />
            </div>
            <div className="text-xs leading-tight truncate">
              <p className="font-bold text-white truncate">{user?.username || 'User'}</p>
              <p className="text-gray-400 truncate text-[10px]">{isConnected ? 'В сети' : 'Оффлайн'}</p>
            </div>
          </div>
          <div className="flex items-center space-x-1">
             <button onClick={logout} className="p-2 hover:bg-[#3f4248] rounded-md transition-colors text-gray-400 hover:text-red-400" title="Выйти">
                <span className="text-lg">🚪</span>
             </button>
          </div>
        </footer>
      </aside>

      {/* ПАНЕЛЬ 3: ОСНОВНОЙ ЧАТ (Центральная) */}
      <main className="flex-1 flex flex-col min-w-0 bg-[#313338]">
        {activeChannel ? (
          <ChatContainer 
            channelId={activeChannel.id} 
            channelName={activeChannel.name} 
          />
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-center p-8">
            <div className="w-20 h-20 bg-[#35363c] rounded-full flex items-center justify-center mb-4">
               <span className="text-4xl text-gray-500">💬</span>
            </div>
            <h2 className="text-2xl font-bold text-white mb-2">Добро пожаловать в Minor!</h2>
            <p className="text-gray-400 max-w-sm">
              Это начало вашего нового сервера. Выберите канал слева или создайте новый, чтобы начать общение.
            </p>
          </div>
        )}
      </main>

      {/* ПАНЕЛЬ 4: СПИСОК УЧАСТНИКОВ (Самая правая) */}
      {activeServer && (
        <MemberList serverId={activeServer.id} />
      )}

      {/* МОДАЛЬНЫЕ ОКНА */}
      <CreateServerModal 
        isOpen={isCreateServerOpen} 
        onClose={() => setIsCreateServerOpen(false)} 
        onServerCreated={handleServerCreated} 
      />
      
      <JoinServerModal 
        isOpen={isJoinServerOpen} 
        onClose={() => setIsJoinServerOpen(false)} 
        onJoined={handleServerCreated} 
      />

      {activeServer && (
        <CreateChannelModal 
          isOpen={isCreateChannelOpen} 
          onClose={() => setIsCreateChannelOpen(false)} 
          serverId={activeServer.id}
          onChannelCreated={handleChannelCreated} 
        />
      )}
    </div>
  );
};

export default MainLayout;