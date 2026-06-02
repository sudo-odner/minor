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
      const response = await api.post(`/api/v1/servers/${serverId}/channels/`, {
        name: name.trim().toLowerCase().replace(/\s+/g, '-'), // форматируем под дискорд-стайл (без пробелов)
        type: type,
      });

      onChannelCreated(response.data);
      setName('');
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Не удалось создать канал');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans">
      <div className="bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in border border-[#1e1f22]">
        
        {/* Шапка */}
        <div className="p-6 pb-4">
          <h3 className="text-xl font-bold text-white">Создать канал</h3>
          <p className="text-gray-400 text-xs mt-1">в текстовых или голосовых каналах</p>
        </div>

        {error && (
          <div className="mx-6 p-2.5 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-6 pt-2 space-y-6">
          {/* Выбор типа канала */}
          <div className="space-y-3">
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Тип канала</label>
            
            {/* Текстовый */}
            <div 
              onClick={() => setType(0)}
              className={`flex items-center p-3 rounded-lg cursor-pointer transition-all ${
                type === 0 ? 'bg-[#3f4248] text-white' : 'bg-[#2b2d31] text-gray-300 hover:bg-[#35373c]'
              }`}
            >
              <span className="text-2xl mr-3 text-gray-400">#</span>
              <div className="text-left">
                <p className="font-semibold text-sm">Текстовый</p>
                <p className="text-gray-400 text-xs">Отправляйте сообщения, изображения, мнения</p>
              </div>
            </div>

            {/* Голосовой */}
            <div 
              onClick={() => setType(1)}
              className={`flex items-center p-3 rounded-lg cursor-pointer transition-all ${
                type === 1 ? 'bg-[#3f4248] text-white' : 'bg-[#2b2d31] text-gray-300 hover:bg-[#35373c]'
              }`}
            >
              <span className="text-2xl mr-3 text-gray-400">🔊</span>
              <div className="text-left">
                <p className="font-semibold text-sm">Голосовой</p>
                <p className="text-gray-400 text-xs">Общайтесь при помощи голоса и видео</p>
              </div>
            </div>
          </div>

          {/* Имя канала */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-gray-300 uppercase tracking-wider">Имя канала</label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-gray-400 text-lg">{type === 0 ? '#' : '🔊'}</span>
              <input
                type="text"
                required
                maxLength={32}
                placeholder="новый-канал"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-[#1e1f22] text-white rounded border border-[#1e1f22] focus:outline-none focus:border-[#5865f2] transition-colors text-sm"
              />
            </div>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-[#2b2d31] -mx-6 -mb-6 p-4 flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-medium text-white hover:underline rounded"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-6 py-2.5 text-sm font-medium bg-[#5865f2] hover:bg-[#4752c4] text-white rounded transition-colors disabled:opacity-50"
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