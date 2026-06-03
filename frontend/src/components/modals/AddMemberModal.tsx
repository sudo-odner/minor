import React, { useState, useEffect } from 'react';
import { api } from '../../api/axios';

interface Friend {
  user_id: string;
  username: string;
  avatar_url: string;
  status: string;
}

interface AddMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  serverId: string;
  onMemberAdded?: () => void;
}

const AddMemberModal: React.FC<AddMemberModalProps> = ({
  isOpen,
  onClose,
  serverId,
  onMemberAdded,
}) => {
  const [friends, setFriends] = useState<Friend[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResult, setSearchResult] = useState<any | null>(null);
  const [loadingFriends, setLoadingFriends] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [addedUserIds, setAddedUserIds] = useState<Record<string, boolean>>({});

  useEffect(() => {
    if (isOpen && serverId) {
      loadFriends();
      setMessage(null);
      setSearchResult(null);
      setSearchQuery('');
    }
  }, [isOpen, serverId]);

  const loadFriends = async () => {
    setLoadingFriends(true);
    try {
      const res = await api.get('/users/friends');
      if (Array.isArray(res.data)) {
        // Фильтруем только тех, кто в статусе "friends" (взаимные друзья)
        const activeFriends = res.data.filter((f: any) => f.status === 'friends');
        setFriends(activeFriends);
      } else {
        setFriends([]);
      }
    } catch (err) {
      console.error('Ошибка при загрузке друзей:', err);
    } finally {
      setLoadingFriends(false);
    }
  };

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;

    setMessage(null);
    setSearchResult(null);

    try {
      const res = await api.get(`/users/search?q=${searchQuery.trim()}`);
      if (res.data) {
        setSearchResult(res.data);
      } else {
        setMessage({ type: 'error', text: 'Пользователь не найден' });
      }
    } catch (err: any) {
      setMessage({ type: 'error', text: 'Пользователь не найден или произошла ошибка' });
    }
  };

  const handleAddMember = async (userId: string, username: string) => {
    setActionLoading(userId);
    setMessage(null);

    try {
      await api.post(`/servers/${serverId}/members`, {
        user_id: userId,
      });

      setAddedUserIds((prev) => ({ ...prev, [userId]: true }));
      setMessage({ type: 'success', text: `Пользователь ${username} успешно добавлен на сервер!` });
      if (onMemberAdded) {
        onMemberAdded();
      }
    } catch (err: any) {
      const errMsg = err.response?.data?.message || err.response?.data?.error;
      if (errMsg === 'already_member' || (typeof errMsg === 'string' && errMsg.includes('already a member'))) {
        setAddedUserIds((prev) => ({ ...prev, [userId]: true }));
        setMessage({ type: 'error', text: `${username} уже является участником этого сервера.` });
      } else {
        setMessage({ type: 'error', text: `Не удалось добавить ${username}: ${errMsg || 'неизвестная ошибка'}` });
      }
    } finally {
      setActionLoading(null);
    }
  };

  if (!isOpen) return null;

  // Фильтруем друзей по локальному поиску (если запрос введен, но форма не отправлена)
  const filteredFriends = friends.filter((f) =>
    f.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 font-sans backdrop-blur-[2px]">
      <div className="bg-white w-full max-w-md rounded-2xl overflow-hidden shadow-2xl animate-fade-in border border-transparent transition-all duration-200">
        
        {/* Шапка */}
        <div className="p-8 pb-4 flex justify-between items-center transition-colors">
          <h3 className="text-2xl font-bold text-gray-800 flex items-center gap-2">
            Участники
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-brand-blue text-lg transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Тело модалки */}
        <div className="p-8 pt-0 space-y-4 max-h-[400px] overflow-y-auto no-scrollbar">
          <p className="text-gray-500 text-sm leading-relaxed">Добавьте друзей на свой сервер, чтобы начать общение.</p>
          
          {message && (
            <div
              className={`p-3 rounded-lg text-sm text-center border transition-colors ${
                message.type === 'success'
                  ? 'bg-green-50 border-green-100 text-green-600'
                  : 'bg-red-50 border-red-100 text-red-600'
              }`}
            >
              {message.type === 'success' ? '✅' : '⚠️'} {message.text}
            </div>
          )}

          {/* Форма поиска */}
          <form onSubmit={handleSearch} className="flex gap-2">
            <input
              type="text"
              placeholder="Поиск по имени/почте..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="flex-1 px-4 py-2.5 bg-gray-50 text-gray-800 rounded-xl border border-gray-200 focus:border-brand-blue focus:outline-none transition-all text-sm placeholder-gray-400"
            />
            <button
              type="submit"
              className="px-5 py-2.5 bg-brand-blue hover:bg-brand-blue-dark text-white text-sm font-bold rounded-xl transition-all shadow-md active:scale-95"
            >
              Найти
            </button>
          </form>

          {/* Результат глобального поиска */}
          {searchResult && (
            <div className="bg-brand-blue-light/20 p-4 rounded-2xl border border-brand-blue-light flex items-center justify-between transition-colors">
              <div className="flex items-center gap-3 min-w-0">
                <div className="w-10 h-10 rounded-full bg-brand-blue flex items-center justify-center text-white font-bold shrink-0 text-sm shadow-sm">
                  {(searchResult.username?.charAt(0) || 'U').toUpperCase()}
                </div>
                <div className="truncate">
                  <h4 className="font-bold text-gray-800 text-sm truncate">{searchResult.username}</h4>
                  <p className="text-xs text-gray-500 truncate">{searchResult.email}</p>
                </div>
              </div>
              <button
                type="button"
                disabled={actionLoading === searchResult.id || addedUserIds[searchResult.id]}
                onClick={() => handleAddMember(searchResult.id, searchResult.username)}
                className={`text-xs px-4 py-2 rounded-lg font-bold transition-all shrink-0 shadow-sm ${
                  addedUserIds[searchResult.id]
                    ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                    : 'bg-[#248046] hover:bg-[#1a6535] text-white active:scale-95'
                }`}
              >
                {addedUserIds[searchResult.id] ? 'Добавлен' : actionLoading === searchResult.id ? '...' : 'Добавить'}
              </button>
            </div>
          )}

          {/* Список друзей */}
          <div className="space-y-2">
            <h4 className="text-xs font-bold text-gray-400 uppercase tracking-widest px-1">Ваши друзья</h4>
            
            {loadingFriends ? (
              <div className="text-center py-4 text-sm text-gray-400">Загрузка...</div>
            ) : filteredFriends.length === 0 ? (
              <div className="text-center py-4 text-sm text-gray-400 italic">
                {searchQuery ? 'Не найдено' : 'Список пуст'}
              </div>
            ) : (
              <div className="space-y-1">
                {filteredFriends.map((friend) => (
                  <div
                    key={friend.user_id}
                    className="flex items-center justify-between p-2 rounded-xl hover:bg-brand-blue-light/30 transition-colors group"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="w-9 h-9 rounded-full bg-brand-blue/10 border border-brand-blue-light flex items-center justify-center text-brand-blue font-bold text-xs shrink-0 transition-colors">
                        {(friend.username?.charAt(0) || 'U').toUpperCase()}
                      </div>
                      <div className="truncate">
                        <span className="font-bold text-gray-600 text-sm block truncate group-hover:text-brand-blue transition-colors">
                          {friend.username}
                        </span>
                      </div>
                    </div>
                    
                    <button
                      type="button"
                      disabled={actionLoading === friend.user_id || addedUserIds[friend.user_id]}
                      onClick={() => handleAddMember(friend.user_id, friend.username)}
                      className={`text-xs px-4 py-2 rounded-lg font-bold transition-all shrink-0 shadow-sm ${
                        addedUserIds[friend.user_id]
                          ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                          : 'bg-brand-blue-light text-brand-blue hover:bg-brand-blue hover:text-white active:scale-95'
                      }`}
                    >
                      {addedUserIds[friend.user_id] ? 'Есть' : actionLoading === friend.user_id ? '...' : 'Добавить'}
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Подвал */}
        <div className="bg-gray-50 p-6 flex justify-end transition-colors duration-200">
          <button
            type="button"
            onClick={onClose}
            className="px-6 py-2.5 text-sm font-bold bg-white hover:bg-gray-100 text-gray-600 border border-gray-200 rounded-xl transition-all shadow-sm active:scale-95"
          >
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddMemberModal;
