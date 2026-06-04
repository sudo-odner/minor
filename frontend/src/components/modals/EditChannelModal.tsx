import React, { useState, useEffect } from 'react';
import { api } from '../../api/axios';

interface EditChannelModalProps {
  isOpen: boolean;
  onClose: () => void;
  serverId: string;
  channel: any;
}

const EditChannelModal: React.FC<EditChannelModalProps> = ({
  isOpen,
  onClose,
  serverId,
  channel,
}) => {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (channel) {
      setName(channel.name || '');
      setError('');
    }
  }, [channel, isOpen]);

  if (!isOpen || !channel) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setLoading(true);
    setError('');

    try {
      // Отправляем запрос на бэкенд в Community Service для переименования
      await api.patch(`/servers/${serverId}/channels/${channel.id}`, {
        name: name.trim().toLowerCase().replace(/\s+/g, '-'),
      });

      onClose();
    } catch (err: any) {
      const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Не удалось обновить канал';
      setError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm(`Вы уверены, что хотите удалить канал #${channel.name}?`)) {
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Отправляем запрос на бэкенд в Community Service для удаления
      await api.delete(`/servers/${serverId}/channels/${channel.id}`);
      onClose();
    } catch (err: any) {
      const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Не удалось удалить канал';
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
          <h3 className="text-2xl font-bold text-gray-800 transition-colors">Настройки канала</h3>
          <p className="text-gray-500 text-sm mt-1 transition-colors">Измените название или удалите канал</p>
        </div>

        {error && (
          <div className="mx-8 p-3 bg-red-50 border border-red-100 text-red-600 text-xs rounded-xl text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-8 pt-4 space-y-6">
          {/* Имя канала */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest px-1">Имя канала</label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-brand-blue opacity-50 text-xl font-light">#</span>
              <input
                type="text"
                required
                maxLength={32}
                placeholder="название-канала"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-gray-50 text-gray-800 rounded-xl border border-gray-200 focus:outline-none focus:border-brand-blue focus:ring-1 focus:ring-brand-blue/10 transition-all text-sm placeholder-gray-400"
              />
            </div>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-gray-50 -mx-8 -mb-8 p-6 flex justify-between items-center transition-colors duration-200 mt-4">
            <button
              type="button"
              onClick={handleDelete}
              disabled={loading}
              className="px-5 py-2.5 text-sm font-bold bg-red-50 hover:bg-red-500 text-red-600 hover:text-white rounded-xl transition-all disabled:opacity-50 active:scale-95 border border-red-100"
            >
              Удалить
            </button>
            <div className="flex space-x-3">
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
                {loading ? '...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditChannelModal;
