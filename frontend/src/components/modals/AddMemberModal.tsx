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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans backdrop-blur-[2px]">
      <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in border border-[#e3e5e8] dark:border-[#1e1f22] transition-colors duration-200">
        
        {/* Шапка */}
        <div className="p-6 pb-4 flex justify-between items-center border-b border-[#e3e5e8] dark:border-[#1e1f22] transition-colors">
          <h3 className="text-xl font-bold text-[#060607] dark:text-white flex items-center gap-2">
            <span>👥</span> Добавить пользователей
          </h3>
          <button
            onClick={onClose}
            className="text-[#4f5660] dark:text-gray-400 hover:text-[#060607] dark:hover:text-white text-lg transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Тело модалки */}
        <div className="p-6 space-y-4 max-h-[400px] overflow-y-auto no-scrollbar">
          {message && (
            <div
              className={`p-3 rounded text-sm text-center border transition-colors ${
                message.type === 'success'
                  ? 'bg-[#248046]/10 border-[#248046]/20 text-[#23a55a]'
                  : 'bg-red-500/10 border-red-500/20 text-red-500 dark:text-red-400'
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
              className="flex-1 px-3 py-2 bg-[#ebedef] dark:bg-[#1e1f22] text-[#060607] dark:text-white rounded border border-transparent focus:border-[#5865f2] focus:outline-none transition-all text-sm placeholder-[#4f5660] dark:placeholder-gray-500"
            />
            <button
              type="submit"
              className="px-4 py-2 bg-[#5865f2] hover:bg-[#4752c4] text-white text-sm font-medium rounded transition-colors shadow-sm"
            >
              Найти
            </button>
          </form>

          {/* Результат глобального поиска */}
          {searchResult && (
            <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] p-3 rounded-lg border border-[#e3e5e8] dark:border-[#3f4248] flex items-center justify-between transition-colors">
              <div className="flex items-center gap-3 min-w-0">
                <div className="w-9 h-9 rounded-full bg-[#5865f2] flex items-center justify-center text-white font-bold shrink-0 text-sm shadow-sm">
                  {(searchResult.username?.charAt(0) || 'U').toUpperCase()}
                </div>
                <div className="truncate">
                  <h4 className="font-bold text-[#060607] dark:text-white text-sm truncate">{searchResult.username}</h4>
                  <p className="text-xs text-[#4f5660] dark:text-gray-400 truncate">{searchResult.email}</p>
                </div>
              </div>
              <button
                type="button"
                disabled={actionLoading === searchResult.id || addedUserIds[searchResult.id]}
                onClick={() => handleAddMember(searchResult.id, searchResult.username)}
                className={`text-xs px-3 py-1.5 rounded font-medium transition-colors shrink-0 shadow-sm ${
                  addedUserIds[searchResult.id]
                    ? 'bg-gray-300 dark:bg-gray-600 text-[#4f5660] dark:text-gray-300 cursor-not-allowed'
                    : 'bg-[#248046] hover:bg-[#1a6535] text-white'
                }`}
              >
                {addedUserIds[searchResult.id] ? 'Добавлен' : actionLoading === searchResult.id ? 'Добавление...' : 'Добавить'}
              </button>
            </div>
          )}

          {/* Список друзей */}
          <div className="space-y-2">
            <h4 className="text-xs font-bold text-[#4f5660] dark:text-gray-400 uppercase tracking-wider transition-colors">Ваши друзья</h4>
            
            {loadingFriends ? (
              <div className="text-center py-4 text-sm text-[#4f5660] dark:text-gray-400">Загрузка друзей...</div>
            ) : filteredFriends.length === 0 ? (
              <div className="text-center py-4 text-sm text-[#4f5660] dark:text-gray-500 italic">
                {searchQuery ? 'Друзья не найдены' : 'Список друзей пуст'}
              </div>
            ) : (
              <div className="space-y-1">
                {filteredFriends.map((friend) => (
                  <div
                    key={friend.user_id}
                    className="flex items-center justify-between p-2 rounded hover:bg-[#f2f3f5] dark:hover:bg-[#2b2d31] transition-colors group"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="w-8 h-8 rounded-full bg-[#d4d7dc] dark:bg-[#3f4248] flex items-center justify-center text-[#4f5660] dark:text-white font-semibold text-xs shrink-0 transition-colors">
                        {(friend.username?.charAt(0) || 'U').toUpperCase()}
                      </div>
                      <div className="truncate">
                        <span className="font-medium text-[#060607] dark:text-white text-sm block truncate group-hover:text-[#060607] dark:group-hover:text-white transition-colors">
                          {friend.username}
                        </span>
                      </div>
                    </div>
                    
                    <button
                      type="button"
                      disabled={actionLoading === friend.user_id || addedUserIds[friend.user_id]}
                      onClick={() => handleAddMember(friend.user_id, friend.username)}
                      className={`text-xs px-3 py-1.5 rounded font-medium transition-colors shrink-0 shadow-sm ${
                        addedUserIds[friend.user_id]
                          ? 'bg-gray-300 dark:bg-gray-600 text-[#4f5660] dark:text-gray-300 cursor-not-allowed'
                          : 'bg-[#248046] hover:bg-[#1a6535] text-white'
                      }`}
                    >
                      {addedUserIds[friend.user_id] ? 'Добавлен' : actionLoading === friend.user_id ? 'Добавление...' : 'Добавить'}
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Подвал */}
        <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] p-4 flex justify-end transition-colors duration-200">
          <button
            type="button"
            onClick={onClose}
            className="px-5 py-2 text-sm font-medium bg-[#ebedef] dark:bg-gray-600 hover:bg-[#d4d7dc] dark:hover:bg-gray-700 text-[#060607] dark:text-white rounded transition-colors shadow-sm"
          >
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddMemberModal;
