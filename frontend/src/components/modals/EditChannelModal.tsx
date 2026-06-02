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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans backdrop-blur-[2px]">
      <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in border border-[#e3e5e8] dark:border-[#1e1f22] transition-colors duration-200">
        
        {/* Шапка */}
        <div className="p-6 pb-4">
          <h3 className="text-xl font-bold text-[#060607] dark:text-white transition-colors">Настройки канала</h3>
          <p className="text-[#4f5660] dark:text-gray-400 text-xs mt-1 transition-colors">Измените название или удалите канал</p>
        </div>

        {error && (
          <div className="mx-6 p-2.5 bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-6 pt-2 space-y-6">
          {/* Имя канала */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-[#4f5660] dark:text-gray-300 uppercase tracking-wider transition-colors">Имя канала</label>
            <div className="relative flex items-center">
              <span className="absolute left-4 text-[#4f5660] dark:text-gray-400 text-lg">#</span>
              <input
                type="text"
                required
                maxLength={32}
                placeholder="название-канала"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-[#ebedef] dark:bg-[#1e1f22] text-[#060607] dark:text-white rounded border border-transparent focus:outline-none focus:border-[#5865f2] transition-all text-sm placeholder-[#4f5660] dark:placeholder-gray-500"
              />
            </div>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] -mx-6 -mb-6 p-4 flex justify-between items-center transition-colors duration-200">
            <button
              type="button"
              onClick={handleDelete}
              disabled={loading}
              className="px-4 py-2.5 text-sm font-medium bg-red-500 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white rounded transition-colors disabled:opacity-50 shadow-sm"
            >
              Удалить канал
            </button>
            <div className="flex space-x-3">
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
                {loading ? 'Сохранение...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditChannelModal;
