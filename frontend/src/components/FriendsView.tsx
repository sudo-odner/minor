import React, { useState, useEffect } from 'react';
import { api } from '../api/axios';
import { useAuth } from '../context/AuthContext';
import { MessageCircle, Ban, Trash2, UserCheck, UserX } from 'lucide-react';

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
    <div className="flex-1 flex flex-col h-full bg-white select-none transition-colors duration-200">
      
      {/* ШАПКА ВКЛАДКИ ДРУЗЕЙ */}
      <header className="h-12 border-b border-brand-blue-light flex items-center px-4 space-x-4 shadow-sm z-10 shrink-0 transition-colors">
        <div className="flex items-center space-x-2 text-brand-blue font-bold border-r border-brand-blue-light pr-4">
          <span className="text-gray-800 text-sm">Друзья</span>
        </div>
        
        {/* Кнопки переключения вкладок */}
        <nav className="flex space-x-2 text-sm font-medium">
          <button 
            onClick={() => setActiveTab('online')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'online' ? 'bg-brand-blue text-white shadow-md' : 'text-gray-500 hover:bg-brand-blue-light hover:text-brand-blue'
            }`}
          >
            В сети
          </button>
          <button 
            onClick={() => setActiveTab('all')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'all' ? 'bg-brand-blue text-white shadow-md' : 'text-gray-500 hover:bg-brand-blue-light hover:text-brand-blue'
            }`}
          >
            Все
          </button>
          <button 
            onClick={() => setActiveTab('pending')}
            className={`px-2 py-1 rounded transition-colors relative ${
              activeTab === 'pending' ? 'bg-brand-blue text-white shadow-md' : 'text-gray-500 hover:bg-brand-blue-light hover:text-brand-blue'
            }`}
          >
            Ожидание
          </button>
          <button 
            onClick={() => setActiveTab('blocked')}
            className={`px-2 py-1 rounded transition-colors ${
              activeTab === 'blocked' ? 'bg-brand-blue text-white shadow-md' : 'text-gray-500 hover:bg-brand-blue-light hover:text-brand-blue'
            }`}
          >
            Заблокированные
          </button>
          <button 
            onClick={() => setActiveTab('add')}
            className={`px-3 py-1 rounded text-white font-bold bg-[#23a55a] hover:bg-[#1a7f45] transition-all shadow-sm ml-2`}
          >
            Добавить
          </button>
        </nav>
      </header>

      {/* ОСНОВНОЙ КОНТЕНТ ВКЛАДОК */}
      <div className="flex-1 overflow-y-auto p-6 bg-brand-bg/30">
        
        {activeTab === 'add' ? (
          /* КЛИЕНТСКИЙ ИНТЕРФЕЙС ДОБАВЛЕНИЯ В ДРУЗЬЯ */
          <div className="max-w-xl animate-fadeIn">
            <h3 className="text-gray-800 font-bold uppercase text-xs mb-2">Добавить друга</h3>
            <p className="text-gray-500 text-sm mb-4 leading-relaxed">
              Вы можете добавить друга по его электронной почте или имени пользователя. Будьте внимательны к регистру букв!
            </p>

            <form onSubmit={handleSearch} className="flex bg-white p-2 rounded-xl border border-gray-200 focus-within:border-brand-blue focus-within:ring-2 focus-within:ring-brand-blue/10 transition-all shadow-sm mb-6">
              <input 
                type="text" 
                placeholder="Введите имя пользователя или email..." 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="bg-transparent border-none outline-none flex-1 text-sm text-gray-800 px-2 placeholder-gray-400"
                disabled={isLoading}
              />
              <button 
                type="submit" 
                className="bg-brand-blue hover:bg-brand-blue-dark text-white text-xs font-bold px-5 py-2 rounded-lg transition-colors"
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
              <div className="bg-white p-5 rounded-2xl border border-gray-100 flex items-center justify-between shadow-xl animate-fadeIn transition-all">
                <div className="flex items-center space-x-4">
                  <div className="w-14 h-14 bg-brand-blue/10 text-brand-blue rounded-full flex items-center justify-center font-bold text-xl border border-brand-blue-light">
                    {(searchResult.username?.charAt(0) || 'U').toUpperCase()}
                  </div>
                  <div>
                    <h4 className="font-bold text-gray-800 text-lg">{searchResult.username}</h4>
                    <p className="text-sm text-gray-500">{searchResult.email}</p>
                  </div>
                </div>
                {(() => {
                  if (searchResult.id === user?.id) {
                    return (
                      <span className="text-gray-400 text-xs font-semibold px-4 py-2 bg-gray-100 rounded-lg">
                        Это вы
                      </span>
                    );
                  }
                  
                  const existingRelation = relationships.find(r => r.user_id === searchResult.id);
                  if (!existingRelation) {
                    return (
                      <button 
                        onClick={() => sendFriendRequest(searchResult.id)}
                        className="bg-[#23a55a] hover:bg-[#1a7f45] text-white text-xs font-bold px-5 py-2.5 rounded-lg transition-all shadow-sm active:scale-95"
                      >
                        Отправить запрос
                      </button>
                    );
                  }

                  if (existingRelation.status === 'friends') {
                    return (
                      <span className="text-brand-blue text-xs font-bold px-4 py-2 bg-brand-blue-light rounded-lg">
                        Вы друзья
                      </span>
                    );
                  }

                  if (existingRelation.status === 'request_sent') {
                    return (
                      <span className="text-gray-500 text-xs font-semibold px-4 py-2 bg-gray-100 rounded-lg">
                        Запрос отправлен
                      </span>
                    );
                  }

                  if (existingRelation.status === 'request_received') {
                    return (
                      <button 
                        onClick={() => handleAnswerRequest(searchResult.id, 'accepted')}
                        className="bg-[#23a55a] hover:bg-[#1a7f45] text-white text-xs font-bold px-5 py-2.5 rounded-lg transition-all shadow-sm"
                      >
                        Принять запрос
                      </button>
                    );
                  }

                  if (existingRelation.status === 'blocked') {
                    return (
                      <button 
                        onClick={() => handleDeleteFriendship(searchResult.id, existingRelation.status)}
                        className="bg-gray-500 hover:bg-gray-600 text-white text-xs font-bold px-5 py-2.5 rounded-lg transition-colors"
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
          <div className="space-y-2 animate-fadeIn max-w-4xl">
            {isLoading ? (
              <div className="text-gray-400 text-center py-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-blue mx-auto mb-4"></div>
                Загрузка данных...
              </div>
            ) : displayList.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-24 text-center text-gray-400 select-none">
                <p className="text-lg font-semibold text-gray-500">Здесь пока пусто</p>
                <p className="text-sm max-w-xs mt-2">
                  Возможно, самое время зайти на вкладку «Добавить друга» и пригласить кого-нибудь.
                </p>
              </div>
            ) : (
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden divide-y divide-gray-50">
                {displayList.map((friend) => (
                  <div 
                    key={friend.user_id} 
                    className="flex items-center justify-between py-4 px-6 hover:bg-brand-blue-light/20 transition-colors group"
                  >
                    <div className="flex items-center space-x-4">
                      {/* Аватар */}
                      <div className="relative shrink-0">
                        <div className="w-12 h-12 bg-brand-blue/10 border border-brand-blue-light rounded-full flex items-center justify-center font-bold text-lg text-brand-blue transition-colors">
                          {(friend.username?.charAt(0) || 'U').toUpperCase()}
                        </div>
                        <div className={`absolute bottom-0 right-0 w-3.5 h-3.5 rounded-full border-2 border-white transition-colors ${
                          friend.isOnline ? 'bg-[#23a55a]' : 'bg-gray-300'
                        }`} />
                      </div>
                      
                      {/* Инфо */}
                      <div>
                        <h4 className="font-bold text-gray-800 text-base group-hover:text-brand-blue transition-colors">
                          {friend.username}
                        </h4>
                        <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">
                          {activeTab === 'pending' 
                            ? (friend.status === 'request_sent' ? 'Исходящий запрос' : 'Входящий запрос')
                            : (friend.status === 'blocked' ? 'В черном списке' : (friend.isOnline ? 'В сети' : 'Офлайн'))
                          }
                        </p>
                      </div>
                    </div>

                    {/* Кнопки действий */}
                    <div className="flex items-center space-x-3">
                      {activeTab === 'pending' && (
                        <>
                          {friend.status === 'request_received' ? (
                            <>
                              <button 
                                onClick={() => handleAnswerRequest(friend.user_id, 'accepted')}
                                className="bg-brand-blue-light text-brand-blue hover:bg-brand-blue hover:text-white p-2.5 rounded-xl transition-all shadow-sm active:scale-95"
                                title="Принять"
                              >
                                <UserCheck size={20} />
                              </button>
                              <button 
                                onClick={() => handleAnswerRequest(friend.user_id, 'deny')}
                                className="bg-gray-50 text-gray-400 hover:bg-brand-blue-light hover:text-brand-blue p-2.5 rounded-xl transition-all shadow-sm active:scale-95"
                                title="Отклонить"
                              >
                                <UserX size={20} />
                              </button>
                            </>
                          ) : (
                            <button 
                              onClick={() => handleDeleteFriendship(friend.user_id, friend.status)}
                              className="bg-gray-100 hover:bg-brand-blue-light text-gray-600 hover:text-brand-blue text-xs font-bold px-4 py-2 rounded-lg transition-colors"
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
                          className="bg-brand-blue-light hover:bg-brand-blue text-brand-blue hover:text-white text-xs font-bold px-4 py-2 rounded-lg transition-all"
                          title="Разблокировать"
                        >
                          Разблокировать
                        </button>
                      )}

                      {(activeTab === 'all' || activeTab === 'online') && (
                        <>
                          <button 
                            onClick={() => alert(`Чат с ${friend.username} в разработке!`)}
                            className="bg-brand-blue-light text-brand-blue hover:bg-brand-blue hover:text-white p-2.5 rounded-xl transition-all shadow-sm active:scale-95"
                            title="Начать чат"
                          >
                            <MessageCircle size={20} />
                          </button>
                          <button 
                            onClick={() => handleBlockUser(friend.user_id)}
                            className="bg-gray-50 text-gray-400 hover:bg-brand-blue-light hover:text-brand-blue p-2.5 rounded-xl transition-all"
                            title="Заблокировать"
                          >
                            <Ban size={20} />
                          </button>
                          <button 
                            onClick={() => handleDeleteFriendship(friend.user_id, friend.status)}
                            className="bg-gray-50 text-gray-400 hover:bg-brand-blue-light hover:text-brand-blue p-2.5 rounded-xl transition-all"
                            title="Удалить из друзей"
                          >
                            <Trash2 size={20} />
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
