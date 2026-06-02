import React, { useState } from 'react';
import { api } from '../../api/axios';

interface CreateServerModalProps {
  isOpen: boolean;
  onClose: () => void;
  onServerCreated: (newServer: any) => void;
}

const CreateServerModal: React.FC<CreateServerModalProps> = ({
  isOpen,
  onClose,
  onServerCreated,
}) => {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setLoading(true);
    setError('');

    try {
      // Endpoint: POST /api/v1/servers/
      const response = await api.post('/servers', {
        name: name.trim(),
        avatar_url: '', // Default empty for now
      });

      onServerCreated(response.data);
      setName('');
      onClose();
    } catch (err: any) {
      const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Не удалось создать сервер';
      setError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans backdrop-blur-[2px]">
      <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in border border-[#e3e5e8] dark:border-[#1e1f22] transition-colors duration-200">
        
        {/* Шапка */}
        <div className="p-6 text-center">
          <h3 className="text-2xl font-bold text-[#060607] dark:text-white transition-colors">Создайте свой сервер</h3>
          <p className="text-[#4f5660] dark:text-gray-400 text-base mt-2 transition-colors">
            Ваш сервер — это место, где вы общаетесь с друзьями. Создайте свой сервер и начните общаться.
          </p>
        </div>

        {error && (
          <div className="mx-6 p-2.5 bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-6 pt-2 space-y-4">
          <div className="space-y-2">
            <label className="text-xs font-bold text-[#4f5660] dark:text-gray-300 uppercase tracking-wider transition-colors">Название сервера</label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Введите название сервера"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3 bg-[#ebedef] dark:bg-[#1e1f22] text-[#060607] dark:text-white rounded border border-transparent focus:outline-none focus:border-[#5865f2] transition-all text-base placeholder-[#4f5660] dark:placeholder-gray-500"
            />
            <p className="text-[10px] text-[#4f5660] dark:text-gray-400 transition-colors">
              Создавая сервер, вы соглашаетесь с Правилами сообщества Minor.
            </p>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-[#f2f3f5] dark:bg-[#2b2d31] -mx-6 -mb-6 p-4 flex justify-between items-center transition-colors duration-200">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-medium text-[#060607] dark:text-white hover:underline transition-colors"
            >
              Назад
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-8 py-2.5 text-sm font-medium bg-[#5865f2] hover:bg-[#4752c4] text-white rounded transition-colors disabled:opacity-50 shadow-sm"
            >
              {loading ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateServerModal;
