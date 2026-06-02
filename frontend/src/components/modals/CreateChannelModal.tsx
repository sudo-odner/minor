import React, { useState } from 'react';
import { api } from '../../api/axios';

interface CreateChannelModalProps {
  isOpen: boolean;
  onClose: () => void;
  serverId: string;
  onChannelCreated: (newChannel: any) => void;
}

const CreateChannelModal: React.FC<CreateChannelModalProps> = ({
  isOpen,
  onClose,
  serverId,
  onChannelCreated,
}) => {
  const [name, setName] = useState('');
  const [type, setType] = useState<number>(0); // 0 - Текстовый, 1 - Голосовой
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setLoading(true);
    setError('');

    try {
      // Отправляем запрос на бэкенд в Community Service
      const response = await api.post(`/servers/${serverId}/channels/`, {
        name: name.trim().toLowerCase().replace(/\s+/g, '-'), // форматируем под дискорд-стайл (без пробелов)
        type: type,
      });

      onChannelCreated(response.data);
      setName('');
      onClose();
    } catch (err: any) {
      const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Не удалось создать канал';
      setError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans backdrop-blur-[2px]">
      <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in border border-[#e3e5e8] dark:border-[#1e1f22] transition-colors duration-200">
        
        {/* Шапка */}
        <div className="p-6 pb-4">
          <h3 className="text-xl font-bold text-[#060607] dark:text-white transition-colors">Создать канал</h3>
          <p className="text-[#4f5660] dark:text-gray-400 text-xs mt-1 transition-colors">в текстовых или голосовых каналах</p>
        </div>

        {error && (
          <div className="mx-6 p-2.5 bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-6 pt-2 space-y-6">
          {/* Выбор типа канала */}
          <div className="space-y-3">
            <label className="text-xs font-bold text-[#4f5660] dark:text-gray-300 uppercase tracking-wider transition-colors">Тип канала</label>
            
            {/* Текстовый */}
            <div 
              onClick={() => setType(0)}
              className={`flex items-center p-3 rounded-lg cursor-pointer transition-all ${
                type === 0 ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white shadow-sm' : 'bg-[#f2f3f5] dark:bg-[#2b2d31] text-[#4f5660] dark:text-gray-300 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c]'
              }`}
            >
              <span className="text-2xl mr-3 text-[#4f5660] dark:text-gray-400">#</span>
              <div className="text-left">
                <p className="font-semibold text-sm">Текстовый</p>
                <p className="text-[#4f5660] dark:text-gray-400 text-xs">Отправляйте сообщения, изображения, мнения</p>
              </div>
            </div>

            {/* Голосовой */}
            <div 
              onClick={() => setType(1)}
              className={`flex items-center p-3 rounded-lg cursor-pointer transition-all ${
                type === 1 ? 'bg-[#d4d7dc] dark:bg-[#3f4248] text-[#060607] dark:text-white shadow-sm' : 'bg-[#f2f3f5] dark:bg-[#2b2d31] text-[#4f5660] dark:text-gray-300 hover:bg-[#e3e5e8] dark:hover:bg-[#35373c]'
              }`}
            >
              <span className="text-2xl mr-3 text-[#4f5660] dark:text-gray-400">🔊</span>
              <div className="text-left">
                <p className="font-semibold text-sm">Голосовой</p>
                <p className="text-[#4f5660] dark:text-gray-400 text-xs">Общайтесь при помощи голоса и видео</p>
              </div>
            </div>
          </div>

          {/* Имя канала */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-[#4f5660] dark:text-gray-300 uppercase tracking-wider transition-colors">Имя канала</label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-[#4f5660] dark:text-gray-400 text-lg">{type === 0 ? '#' : '🔊'}</span>
              <input
                type="text"
                required
                maxLength={32}
                placeholder="новый-канал"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-[#ebedef] dark:bg-[#1e1f22] text-[#060607] dark:text-white rounded border border-transparent focus:outline-none focus:border-[#5865f2] transition-all text-sm placeholder-[#4f5660] dark:placeholder-gray-500"
              />
            </div>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] -mx-6 -mb-6 p-4 flex justify-end space-x-3 transition-colors duration-200">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-medium text-[#060607] dark:text-white hover:underline rounded transition-colors"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-6 py-2.5 text-sm font-medium bg-[#5865f2] hover:bg-[#4752c4] text-white rounded transition-colors disabled:opacity-50 shadow-sm"
            >
              {loading ? 'Создание...' : 'Создать канал'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateChannelModal;