import React, { useState, useEffect } from 'react';
import { api } from '../api/axios';
import { useAuth } from '../context/AuthContext';

interface Friend {
  user_id: string;
  username: string;
  avatar_url: string;
  status: string; // "friends" | "request_sent" | "request_received" | "blocked"
  isOnline?: boolean;
}

interface FriendsViewProps {
  onlineUsers?: Record<string, boolean>;
}

const FriendsView: React.FC<FriendsViewProps> = ({ onlineUsers = {} }) => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'online' | 'all' | 'pending' | 'blocked' | 'add'>('online');
  const [relationships, setRelationships] = useState<Friend[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResult, setSearchResult] = useState<any | null>(null);
  const [searchError, setSearchError] = useState('');
  const [searchSuccess, setSearchSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // 1. Загрузка списка всех связей
  const loadRelationships = async () => {
    setIsLoading(true);
    try {
      const res = await api.get('/users/friends');
      const list = Array.isArray(res.data) ? res.data : [];
      setRelationships(list.map((f: any) => ({ ...f, isOnline: f.is_online ?? false })));
    } catch (err) {
      console.error('Ошибка загрузки связей:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadRelationships();
  }, []);

  // Apply real-time presence updates from parent
  useEffect(() => {
    if (Object.keys(onlineUsers).length === 0) return;
    setRelationships(prev => prev.map(f => {
      if (f.user_id in onlineUsers) {
        return { ...f, isOnline: onlineUsers[f.user_id] };
      }
      return f;
    }));
  }, [onlineUsers]);

  useEffect(() => {
    loadRelationships();
    if (activeTab === 'add') {
      setSearchResult(null);
      setSearchError('');
      setSearchSuccess('');
    }
  }, [activeTab]);

  // 2. Поиск пользователя по логину или почте
  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;

    setSearchResult(null);
    setSearchError('');
    setSearchSuccess('');
    setIsLoading(true);

    try {
      const res = await api.get(`/users/search?q=${searchQuery.trim()}`);
      if (res.data) {
        setSearchResult(res.data);
      } else {
        setSearchError('Пользователь не найден');
      }
    } catch (err: any) {
      if (err.response?.status === 404) {
        setSearchError('Пользователь не найден');
      } else {
        setSearchError('Произошла ошибка при поиске пользователя');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // 3. Отправка запроса в друзья
  const sendFriendRequest = async (friendId: string) => {
    try {
      await api.post(`/users/friends/requests/${friendId}`);
      setSearchSuccess('Заявка в друзья успешно отправлена!');
      loadRelationships();
    } catch (err: any) {
      setSearchError(err.response?.data?.message || 'Не удалось отправить запрос в друзья');
    }
  };

  // 4. Управление запросами (Принять / Отклонить)
  const handleAnswerRequest = async (friendId: string, status: 'accepted' | 'deny') => {
    try {
      await api.patch(`/users/friends/requests/${friendId}`, { status });
      loadRelationships();
    } catch (err) {
      alert('Ошибка обработки запроса');
    }
  };

  // 5. Удаление из друзей / Отмена заявки
  const handleDeleteFriendship = async (friendId: string, status?: string) => {
    let confirmMsg = 'Вы уверены, что хотите удалить этого пользователя из друзей?';
    if (status === 'request_sent') {
      confirmMsg = 'Вы уверены, что хотите отменить исходящую заявку?';
    } else if (status === 'blocked') {
      confirmMsg = 'Вы уверены, что хотите разблокировать этого пользователя?';
    }
    if (!window.confirm(confirmMsg)) return;
    try {
      await api.delete(`/users/friends/${friendId}`);
      loadRelationships();
    } catch (err) {
      alert('Ошибка удаления связи');
    }
  };

  // 6. Блокировка пользователя
  const handleBlockUser = async (friendId: string) => {
    try {
      await api.post(`/users/friends/block/${friendId}`);
      loadRelationships();
    } catch (err) {
      alert('Ошибка блокировки пользователя');
    }
  };

  // Фильтр списка друзей для текущей вкладки
  const displayList = relationships.filter(friend => {
    if (activeTab === 'online') return friend.status === 'friends' && friend.isOnline;
    if (activeTab === 'all') return friend.status === 'friends';
    if (activeTab === 'pending') return friend.status === 'request_sent' || friend.status === 'request_received';
    if (activeTab === 'blocked') return friend.status === 'blocked';
    return false;
  });

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#313338] select-none transition-colors duration-200">
      
      {/* ШАПКА ВКЛАДКИ ДРУЗЕЙ */}
      <header className="h-12 border-b border-[#e3e5e8] dark:border-[#1e1f22] flex items-center px-4 space-x-4 shadow-sm z-10 shrink-0 transition-colors">
        <div className="flex items-center space-x-2 text-[#4f5660] dark:text-gray-400 font-bold border-r border-[#e3e5e8] dark:border-[#3f4147] pr-4">
          <span className="text-xl">👥</span>
          <span className="text-[#060607] dark:text-white text-sm">Друзья</span>
        </div>
        
        {/* Кнопки переключения вкладок */}
        <nav className="flex space-x-2 text-sm font-medium">
          <button 
            onClick={() => setActiveTab('online')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'online' ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
            }`}
          >
            В сети
          </button>
          <button 
            onClick={() => setActiveTab('all')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'all' ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
            }`}
          >
            Все
          </button>
          <button 
            onClick={() => setActiveTab('pending')}
            className={`px-2 py-1 rounded transition-colors relative ${
              activeTab === 'pending' ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
            }`}
          >
            Ожидание
          </button>
          <button 
            onClick={() => setActiveTab('blocked')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'blocked' ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white' : 'text-[#4f5660] dark:text-gray-400 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] hover:text-[#060607] dark:hover:text-gray-200'
            }`}
          >
            Заблокированные
          </button>
          <button 
            onClick={() => setActiveTab('add')}
            className={`px-2.5 py-1 rounded text-white font-bold bg-[#23a55a] hover:bg-[#1a7f45] transition-colors`}
          >
            Добавить друга
          </button>
        </nav>
      </header>

      {/* ОСНОВНОЙ КОНТЕНТ ВКЛАДОК */}
      <div className="flex-1 overflow-y-auto p-6">
        
        {activeTab === 'add' ? (
          /* КЛИЕНТСКИЙ ИНТЕРФЕЙС ДОБАВЛЕНИЯ В ДРУЗЬЯ */
          <div className="max-w-xl animate-fadeIn">
            <h3 className="text-[#060607] dark:text-white font-bold uppercase text-xs mb-2">Добавить друга</h3>
            <p className="text-[#4f5660] dark:text-gray-400 text-sm mb-4 leading-relaxed">
              Вы можете добавить друга по его электронной почте или имени пользователя. Будьте внимательны к регистру букв!
            </p>

            <form onSubmit={handleSearch} className="flex bg-[#ebedef] dark:bg-[#1e1f22] p-3 rounded-lg border border-transparent focus-within:border-[#5865f2] transition-colors shadow-inner mb-6">
              <input 
                type="text" 
                placeholder="Введите имя пользователя или email..." 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="bg-transparent border-none outline-none flex-1 text-sm text-[#060607] dark:text-gray-200 placeholder-[#4f5660] dark:placeholder-gray-500"
                disabled={isLoading}
              />
              <button 
                type="submit" 
                className="bg-[#5865f2] hover:bg-[#4752c4] text-white text-xs font-bold px-4 py-1.5 rounded transition-colors"
                disabled={isLoading || !searchQuery.trim()}
              >
                {isLoading ? 'Поиск...' : 'Найти'}
              </button>
            </form>

            {/* Ошибки или успехи */}
            {searchError && (
              <div className="text-red-500 text-sm mb-4 bg-red-500/10 p-3 rounded border border-red-500/20">
                ⚠️ {searchError}
              </div>
            )}

            {searchSuccess && (
              <div className="text-green-500 text-sm mb-4 bg-green-500/10 p-3 rounded border border-green-500/20">
                ✅ {searchSuccess}
              </div>
            )}

            {/* Карточка найденного пользователя */}
            {searchResult && (
              <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] p-4 rounded-xl border border-[#e3e5e8] dark:border-[#3f4248] flex items-center justify-between shadow-lg animate-fadeIn transition-colors">
                <div className="flex items-center space-x-3">
                  <div className="w-12 h-12 bg-[#5865f2] rounded-full flex items-center justify-center font-bold text-lg text-white">
                    {(searchResult.username?.charAt(0) || 'U').toUpperCase()}
                  </div>
                  <div>
                    <h4 className="font-bold text-[#060607] dark:text-white text-base">{searchResult.username}</h4>
                    <p className="text-xs text-[#4f5660] dark:text-gray-400">{searchResult.email}</p>
                  </div>
                </div>
                {(() => {
                  if (searchResult.id === user?.id) {
                    return (
                      <span className="text-[#4f5660] dark:text-gray-400 text-xs font-semibold px-4 py-2 bg-[#d4d7dc] dark:bg-[#3f4248] rounded-lg">
                        Это вы
                      </span>
                    );
                  }
                  
                  const existingRelation = relationships.find(r => r.user_id === searchResult.id);
                  if (!existingRelation) {
                    return (
                      <button 
                        onClick={() => sendFriendRequest(searchResult.id)}
                        className="bg-[#23a55a] hover:bg-[#1a7f45] text-white text-xs font-bold px-4 py-2 rounded-lg transition-colors"
                      >
                        Отправить запрос
                      </button>
                    );
                  }

                  if (existingRelation.status === 'friends') {
                    return (
                      <span className="text-[#4f5660] dark:text-gray-400 text-xs font-semibold px-4 py-2 bg-[#d4d7dc] dark:bg-[#3f4248] rounded-lg">
                        Вы друзья
                      </span>
                    );
                  }

                  if (existingRelation.status === 'request_sent') {
                    return (
                      <span className="text-[#4f5660] dark:text-gray-400 text-xs font-semibold px-4 py-2 bg-[#d4d7dc] dark:bg-[#3f4248] rounded-lg">
                        Запрос отправлен
                      </span>
                    );
                  }

                  if (existingRelation.status === 'request_received') {
                    return (
                      <button 
                        onClick={() => handleAnswerRequest(searchResult.id, 'accepted')}
                        className="bg-[#23a55a] hover:bg-[#1a7f45] text-white text-xs font-bold px-4 py-2 rounded-lg transition-colors"
                      >
                        Принять запрос
                      </button>
                    );
                  }

                  if (existingRelation.status === 'blocked') {
                    return (
                      <button 
                        onClick={() => handleDeleteFriendship(searchResult.id, existingRelation.status)}
                        className="bg-gray-500 hover:bg-gray-400 text-white text-xs font-bold px-4 py-2 rounded-lg transition-colors"
                      >
                        Разблокировать
                      </button>
                    );
                  }

                  return null;
                })()}
              </div>
            )}
          </div>
        ) : (
          /* СПИСОК СВЯЗЕЙ (ДРУЗЬЯ / ЗАПРОСЫ / ЧС) */
          <div className="space-y-2 animate-fadeIn">
            {isLoading ? (
              <div className="text-[#4f5660] dark:text-gray-400 text-center py-8">Загрузка данных...</div>
            ) : displayList.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center text-[#4f5660] dark:text-gray-500 select-none">
                <span className="text-6xl mb-4 opacity-50">🏜️</span>
                <p className="text-base font-medium">Никого нет. Совсем.</p>
                <p className="text-xs max-w-xs mt-1">
                  Возможно, самое время зайти на вкладку «Добавить друга» и пригласить кого-нибудь.
                </p>
              </div>
            ) : (
              <div className="divide-y divide-[#e3e5e8] dark:divide-[#35363c] border-t border-b border-[#e3e5e8] dark:border-[#35363c] transition-colors">
                {displayList.map((friend) => (
                  <div 
                    key={friend.user_id} 
                    className="flex items-center justify-between py-3 hover:bg-[#f2f3f5] dark:hover:bg-[#35373c] -mx-6 px-6 transition-colors group"
                  >
                    <div className="flex items-center space-x-3">
                      {/* Аватар */}
                      <div className="relative shrink-0">
                        <div className="w-10 h-10 bg-[#e3e5e8] dark:bg-[#313338] border border-[#d4d7dc] dark:border-[#3f4248] rounded-full flex items-center justify-center font-bold text-sm text-[#4f5660] dark:text-gray-300 transition-colors">
                          {(friend.username?.charAt(0) || 'U').toUpperCase()}
                        </div>
                        <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-[#313338] transition-colors ${
                          friend.isOnline ? 'bg-[#23a55a]' : 'bg-gray-500'
                        }`} />
                      </div>
                      
                      {/* Инфо */}
                      <div>
                        <h4 className="font-bold text-[#060607] dark:text-white text-sm group-hover:text-[#060607] dark:group-hover:text-white transition-colors">
                          {friend.username}
                        </h4>
                        <p className="text-xs text-[#4f5660] dark:text-gray-400">
                          {activeTab === 'pending' 
                            ? (friend.status === 'request_sent' ? 'Исходящий запрос' : 'Входящий запрос')
                            : (friend.status === 'blocked' ? 'В черном списке' : (friend.isOnline ? 'В сети' : 'Офлайн'))
                          }
                        </p>
                      </div>
                    </div>

                    {/* Кнопки действий */}
                    <div className="flex items-center space-x-2">
                      {activeTab === 'pending' && (
                        <>
                          {friend.status === 'request_received' ? (
                            <>
                              <button 
                                onClick={() => handleAnswerRequest(friend.user_id, 'accepted')}
                                className="bg-[#23a55a] hover:bg-[#1a7f45] text-white p-2 rounded-full transition-colors shadow-sm"
                                title="Принять"
                              >
                                <span>✔️</span>
                              </button>
                              <button 
                                onClick={() => handleAnswerRequest(friend.user_id, 'deny')}
                                className="bg-red-500 hover:bg-red-600 text-white p-2 rounded-full transition-colors shadow-sm"
                                title="Отклонить"
                              >
                                <span>❌</span>
                              </button>
                            </>
                          ) : (
                            <button 
                              onClick={() => handleDeleteFriendship(friend.user_id, friend.status)}
                              className="bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-[#060607] dark:text-white text-xs px-3 py-1.5 rounded transition-colors"
                              title="Отменить заявку"
                            >
                              Отменить
                            </button>
                          )}
                        </>
                      )}

                      {activeTab === 'blocked' && (
                        <button 
                          onClick={() => handleDeleteFriendship(friend.user_id, friend.status)}
                          className="bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-[#060607] dark:text-white text-xs px-3 py-1.5 rounded transition-colors"
                          title="Разблокировать"
                        >
                          Разблокировать
                        </button>
                      )}

                      {(activeTab === 'all' || activeTab === 'online') && (
                        <>
                          <button 
                            onClick={() => alert(`Чат с ${friend.username} в разработке!`)}
                            className="bg-[#f2f3f5] dark:bg-[#2b2d31] hover:bg-[#e3e5e8] dark:hover:bg-[#1e1f22] p-2.5 rounded-full transition-colors text-[#4f5660] dark:text-gray-300 hover:text-[#060607] dark:hover:text-white shadow-sm"
                            title="Начать чат"
                          >
                            <span>💬</span>
                          </button>
                          <button 
                            onClick={() => handleBlockUser(friend.user_id)}
                            className="bg-[#f2f3f5] dark:bg-[#2b2d31] hover:bg-red-500/10 p-2.5 rounded-full transition-colors text-[#4f5660] dark:text-gray-400 hover:text-red-500 shadow-sm"
                            title="Заблокировать"
                          >
                            <span>🚫</span>
                          </button>
                          <button 
                            onClick={() => handleDeleteFriendship(friend.user_id, friend.status)}
                            className="bg-[#f2f3f5] dark:bg-[#2b2d31] hover:bg-red-500 dark:hover:bg-red-600 p-2.5 rounded-full transition-colors text-[#4f5660] dark:text-gray-400 hover:text-white shadow-sm"
                            title="Удалить из друзей"
                          >
                            <span>🗑️</span>
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

      </div>
    </div>
  );
};

export default FriendsView;
