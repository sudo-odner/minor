import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useSocket } from '../context/SocketContext';
import { api } from '../api/axios';

// Импорт компонентов
import CreateServerModal from '../components/modals/CreateServerModal';
import JoinServerModal from '../components/modals/JoinServerModal';
import CreateChannelModal from '../components/modals/CreateChannelModal';
import EditChannelModal from '../components/modals/EditChannelModal';
import AddMemberModal from '../components/modals/AddMemberModal';
import MemberList from '../components/server/MemberList';
import ChatContainer from '../components/chat/ChatContainer';
import FriendsView from '../components/FriendsView';
import { useTheme } from '../context/ThemeContext';

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout = ({ children }: MainLayoutProps) => {
  const { logout, user, accessToken } = useAuth();
  const { socket, isConnected } = useSocket();
  const { theme, toggleTheme } = useTheme();

  // Списки данных
  const [servers, setServers] = useState<any[]>([]);
  const [channels, setChannels] = useState<any[]>([]);
  const [memberListVersion, setMemberListVersion] = useState(0);
  // Map of userId -> isOnline for real-time presence
  const [onlineUsers, setOnlineUsers] = useState<Record<string, boolean>>({});
  
  // Состояния выбора
  const [activeServer, setActiveServer] = useState<any | null>(null);
  const [activeChannel, setActiveChannel] = useState<any | null>(null);

  // Состояния модалок
  const [isCreateServerOpen, setIsCreateServerOpen] = useState(false);
  const [isJoinServerOpen, setIsJoinServerOpen] = useState(false);
  const [isCreateChannelOpen, setIsCreateChannelOpen] = useState(false);
  const [isEditChannelOpen, setIsEditChannelOpen] = useState(false);
  const [selectedChannel, setSelectedChannel] = useState<any | null>(null);
  const [isAddMemberOpen, setIsAddMemberOpen] = useState(false);

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

  // 3. Прослушивание WebSocket для событий каналов и участников сообщества в реальном времени
  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.op === 1) {
          const data = payload.d;
          // Presence updates — apply globally regardless of active server
          if (payload.t === 'PRESENCE_UPDATE' && data?.user_id) {
            const isOnline = data.status === 1;
            setOnlineUsers(prev => ({ ...prev, [data.user_id]: isOnline }));
          }
          if (data && data.server_id === activeServer?.id) {
            switch (payload.t) {
              case 'CHANNEL_CREATE':
                setChannels((prev: any[]) => {
                  if (prev.some(ch => ch.id === data.channel.id)) return prev;
                  return [...prev, data.channel];
                });
                break;
              case 'CHANNEL_UPDATE':
                setChannels((prev: any[]) => prev.map(ch => ch.id === data.channel.id ? data.channel : ch));
                break;
              case 'CHANNEL_DELETE':
                setChannels((prev: any[]) => prev.filter(ch => ch.id !== data.channel_id));
                setActiveChannel((prev: any) => prev && prev.id === data.channel_id ? null : prev);
                break;
              case 'MEMBER_ADD':
              case 'MEMBER_REMOVE':
                setMemberListVersion((prev: number) => prev + 1);
                break;
            }
          }
        }
      } catch (err) {
        console.error('Error processing websocket message in MainLayout:', err);
      }
    };

    socket.addEventListener('message', handleMessage);
    return () => socket.removeEventListener('message', handleMessage);
  }, [socket, activeServer]);

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
    <div className="flex h-screen w-full bg-white dark:bg-[#313338] text-[#060607] dark:text-white overflow-hidden font-sans transition-colors duration-200">
      
      {/* ПАНЕЛЬ 1: СПИСОК СЕРВЕРОВ (Самая левая) */}
      <nav className="w-[72px] bg-[#e3e5e8] dark:bg-[#1e1f22] flex flex-col items-center py-3 space-y-2 overflow-y-auto no-scrollbar transition-colors duration-200">
        {/* Кнопка "Личные сообщения" */}
        <div 
          onClick={() => setActiveServer(null)}
          className={`w-12 h-12 rounded-full flex items-center justify-center cursor-pointer transition-all duration-200 ${
            !activeServer ? 'bg-[#5865f2] rounded-xl text-white' : 'bg-white dark:bg-[#313338] text-[#060607] dark:text-white hover:bg-[#5865f2] hover:text-white hover:rounded-xl'
          }`}
        >
          <span className="text-xl font-bold">M</span>
        </div>

        <div className="w-8 h-[2px] bg-[#d1d3d6] dark:bg-[#35363c] rounded-full mx-auto my-1 transition-colors" />

        {/* Список иконок серверов */}
        {servers.map((server) => (
          <div
            key={server.id}
            onClick={() => setActiveServer(server)}
            className={`relative flex items-center group mb-2`}
          >
            {/* Полоска индикатор слева */}
            <div className={`absolute -left-3 w-1 bg-[#060607] dark:bg-white rounded-r-full transition-all duration-200 ${
              activeServer?.id === server.id ? 'h-10' : 'h-2 scale-0 group-hover:scale-100 group-hover:h-5'
            }`} />
            
            <div className={`w-12 h-12 flex items-center justify-center cursor-pointer transition-all duration-200 shadow-lg ${
              activeServer?.id === server.id ? 'bg-[#5865f2] rounded-xl text-white' : 'bg-white dark:bg-[#313338] text-[#060607] dark:text-white rounded-full hover:rounded-xl hover:bg-[#5865f2] hover:text-white'
            }`}>
              <span className="text-sm font-bold uppercase">{(server.name || '??').substring(0, 2)}</span>
            </div>
          </div>
        ))}

        {/* Кнопки действий */}
        <button 
          onClick={() => setIsCreateServerOpen(true)}
          className="w-12 h-12 bg-white dark:bg-[#313338] rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl hover:bg-[#23a55a] transition-all text-[#23a55a] hover:text-white"
        >
          <span className="text-2xl font-light">+</span>
        </button>

        <button 
          onClick={() => setIsJoinServerOpen(true)}
          className="w-12 h-12 bg-white dark:bg-[#313338] rounded-full flex items-center justify-center cursor-pointer hover:rounded-xl hover:bg-[#5865f2] transition-all text-gray-400 hover:text-white"
        >
          <span className="text-xl">🧭</span>
        </button>
      </nav>

      {/* ПАНЕЛЬ 2: СПИСОК КАНАЛОВ (Средняя) */}
      <aside className="w-60 bg-[#f2f3f5] dark:bg-[#2b2d31] flex flex-col shrink-0 transition-colors duration-200">
        <header className="h-12 shadow-sm flex items-center justify-between px-4 font-bold border-b border-[#e3e5e8] dark:border-[#1e1f22] z-10">
          <span className="truncate">{activeServer ? activeServer.name : "Личные сообщения"}</span>
          {activeServer && (
            <button
              onClick={() => setIsAddMemberOpen(true)}
              className="text-xs bg-[#248046] hover:bg-[#1a6535] text-white px-2 py-1 rounded transition-colors font-medium ml-2 shrink-0"
              title="Добавить участников"
            >
              + Добавить
            </button>
          )}
        </header>

        <div className="flex-1 overflow-y-auto p-2">
          {activeServer ? (
            <>
              <div className="flex items-center justify-between text-xs font-bold text-[#4f5660] dark:text-gray-400 px-2 mt-4 mb-2 uppercase tracking-wider">
                <span>Текстовые каналы</span>
                <button onClick={() => setIsCreateChannelOpen(true)} className="hover:text-[#060607] dark:hover:text-white text-xl leading-none">+</button>
              </div>
              
              <div className="space-y-0.5">
                {channels.map((channel) => (
                  <div
                    key={channel.id}
                    onClick={() => setActiveChannel(channel)}
                    className={`group flex items-center justify-between px-2 py-1.5 rounded cursor-pointer transition-colors ${
                      activeChannel?.id === channel.id ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
                    }`}
                  >
                    <div className="flex items-center min-w-0">
                      <span className="text-[#4f5660] dark:text-gray-500 text-xl mr-1.5">#</span>
                      <span className="truncate font-medium">{channel.name}</span>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedChannel(channel);
                        setIsEditChannelOpen(true);
                      }}
                      className="opacity-0 group-hover:opacity-100 hover:text-[#060607] dark:hover:text-white transition-opacity ml-2 text-sm text-[#4f5660] dark:text-gray-400"
                      title="Настройки канала"
                    >
                      ⚙️
                    </button>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <>
              <div 
                onClick={() => setActiveChannel(null)}
                className={`flex items-center px-3 py-2 rounded-lg cursor-pointer transition-colors mb-4 ${
                  !activeChannel ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
                }`}
              >
                <span className="text-xl mr-2.5">👥</span>
                <span className="font-semibold text-sm">Друзья</span>
              </div>
              
              <div className="text-xs font-bold text-[#4f5660] dark:text-gray-400 px-2 uppercase tracking-wider mb-2">
                Личные сообщения
              </div>
              <div className="space-y-0.5">
                <div className="px-2 py-1.5 text-xs text-[#4f5660] dark:text-gray-500 italic">
                  Нажмите на друга для начала личной переписки.
                </div>
              </div>
            </>
          )}
        </div>

        {/* ПАНЕЛЬ ПРОФИЛЯ (Футер средней колонки) */}
        <footer className="bg-[#ebedef] dark:bg-[#232428] h-[52px] px-2 flex items-center justify-between shrink-0 transition-colors duration-200">
          <div className="flex items-center space-x-2 min-w-0">
            <div className="relative shrink-0">
              <div className="w-8 h-8 bg-[#5865f2] rounded-full flex items-center justify-center font-bold text-xs text-white">
                  {(user?.username?.charAt(0) || 'U').toUpperCase()}
              </div>
              <div className={`absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-[3px] border-[#ebedef] dark:border-[#232428] ${
                isConnected ? 'bg-[#23a55a]' : 'bg-[#80848e]'
              }`} />
            </div>
            <div className="text-xs leading-tight truncate">
              <p className="font-bold text-[#060607] dark:text-white truncate">{user?.username || 'User'}</p>
              <p className="text-[#4f5660] dark:text-gray-400 truncate text-[10px]">{isConnected ? 'В сети' : 'Оффлайн'}</p>
            </div>
          </div>
          <div className="flex items-center space-x-1">
             <button 
                onClick={toggleTheme} 
                className="p-2 hover:bg-[#d4d7dc] dark:hover:bg-[#3f4248] rounded-md transition-colors text-[#4f5660] dark:text-gray-400 hover:text-[#060607] dark:hover:text-white"
                title={theme === 'light' ? "Тёмная тема" : "Светлая тема"}
             >
                <span className="text-lg">{theme === 'light' ? '🌙' : '☀️'}</span>
             </button>
             <button onClick={logout} className="p-2 hover:bg-[#d4d7dc] dark:hover:bg-[#3f4248] rounded-md transition-colors text-[#4f5660] dark:text-gray-400 hover:text-red-500" title="Выйти">
                <span className="text-lg">🚪</span>
             </button>
          </div>
        </footer>
      </aside>

      {/* ПАНЕЛЬ 3: ОСНОВНОЙ ЧАТ (Центральная) */}
      <main className="flex-1 flex flex-col min-w-0 bg-white dark:bg-[#313338] transition-colors duration-200">
        {activeChannel ? (
          <ChatContainer 
            channelId={activeChannel.id} 
            channelName={activeChannel.name} 
          />
        ) : !activeServer ? (
          <FriendsView onlineUsers={onlineUsers} />
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-center p-8">
            <div className="w-20 h-20 bg-[#f2f3f5] dark:bg-[#35363c] rounded-full flex items-center justify-center mb-4 transition-colors">
               <span className="text-4xl text-[#4f5660] dark:text-gray-500">💬</span>
            </div>
            <h2 className="text-2xl font-bold text-[#060607] dark:text-white mb-2 transition-colors">Добро пожаловать в Minor!</h2>
            <p className="text-[#4f5660] dark:text-gray-400 max-w-sm transition-colors">
              Это начало вашего нового сервера. Выберите канал слева или создайте новый, чтобы начать общение.
            </p>
          </div>
        )}
      </main>

      {/* ПАНЕЛЬ 4: СПИСОК УЧАСТНИКОВ (Самая правая) */}
      {activeServer && (
        <MemberList serverId={activeServer.id} onlineUsers={onlineUsers} key={`${activeServer.id}-${memberListVersion}`} />
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

      {activeServer && selectedChannel && (
        <EditChannelModal
          isOpen={isEditChannelOpen}
          onClose={() => {
            setIsEditChannelOpen(false);
            setSelectedChannel(null);
          }}
          serverId={activeServer.id}
          channel={selectedChannel}
        />
      )}

      {activeServer && (
        <AddMemberModal
          isOpen={isAddMemberOpen}
          onClose={() => setIsAddMemberOpen(false)}
          serverId={activeServer.id}
          onMemberAdded={() => setMemberListVersion(prev => prev + 1)}
        />
      )}
    </div>
  );
};

export default MainLayout;