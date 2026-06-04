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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 font-sans backdrop-blur-[2px]">
      <div className="bg-white w-full max-w-md rounded-2xl overflow-hidden shadow-2xl animate-fade-in border border-transparent transition-all duration-200">
        
        {/* Шапка */}
        <div className="p-8 pb-4">
          <h3 className="text-2xl font-bold text-gray-800 transition-colors">Создать канал</h3>
          <p className="text-gray-500 text-sm mt-1 transition-colors">в текстовых или голосовых каналах</p>
        </div>

        {error && (
          <div className="mx-8 p-3 bg-red-50 border border-red-100 text-red-600 text-xs rounded-xl text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-8 pt-4 space-y-6">
          {/* Выбор типа канала */}
          <div className="space-y-3">
            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest px-1">Тип канала</label>
            
            {/* Текстовый */}
            <div 
              onClick={() => setType(0)}
              className={`flex items-center p-4 rounded-2xl cursor-pointer transition-all border ${
                type === 0 ? 'bg-brand-blue-light/50 border-brand-blue-light text-brand-blue shadow-sm' : 'bg-gray-50 border-transparent text-gray-500 hover:bg-gray-100'
              }`}
            >
              <span className="text-3xl mr-4 opacity-50">#</span>
              <div className="text-left">
                <p className="font-bold text-sm">Текстовый</p>
                <p className="opacity-70 text-xs mt-0.5">Отправляйте сообщения, изображения, мнения</p>
              </div>
            </div>

            {/* Голосовой */}
            <div 
              onClick={() => setType(1)}
              className={`flex items-center p-4 rounded-2xl cursor-pointer transition-all border ${
                type === 1 ? 'bg-brand-blue-light/50 border-brand-blue-light text-brand-blue shadow-sm' : 'bg-gray-50 border-transparent text-gray-500 hover:bg-gray-100'
              }`}
            >
              <span className="text-3xl mr-4 opacity-50">🔊</span>
              <div className="text-left">
                <p className="font-bold text-sm">Голосовой</p>
                <p className="opacity-70 text-xs mt-0.5">Общайтесь при помощи голоса и видео</p>
              </div>
            </div>
          </div>

          {/* Имя канала */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest px-1">Имя канала</label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-brand-blue opacity-50 text-xl font-light">{type === 0 ? '#' : '🔊'}</span>
              <input
                type="text"
                required
                maxLength={32}
                placeholder="новый-канал"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-gray-50 text-gray-800 rounded-xl border border-gray-200 focus:outline-none focus:border-brand-blue focus:ring-1 focus:ring-brand-blue/10 transition-all text-sm placeholder-gray-400"
              />
            </div>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-gray-50 -mx-8 -mb-8 p-6 flex justify-end space-x-3 transition-colors duration-200">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-bold text-gray-500 hover:text-gray-800 transition-colors"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-8 py-2.5 text-sm font-bold bg-brand-blue hover:bg-brand-blue-dark text-white rounded-xl transition-all disabled:opacity-50 shadow-md active:scale-95"
            >
              {loading ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateChannelModal;