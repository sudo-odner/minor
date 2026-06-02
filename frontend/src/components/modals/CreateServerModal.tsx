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
      const response = await api.post('/api/v1/servers', {
        name: name.trim(),
        avatar_url: '', // Default empty for now
      });

      onServerCreated(response.data);
      setName('');
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Не удалось создать сервер');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans">
      <div className="bg-[#ffffff] w-full max-w-md rounded-lg overflow-hidden shadow-2xl animate-fade-in">
        
        {/* Шапка */}
        <div className="p-6 text-center">
          <h3 className="text-2xl font-bold text-[#060607]">Создайте свой сервер</h3>
          <p className="text-[#4e5058] text-base mt-2">
            Ваш сервер — это место, где вы общаетесь с друзьями. Создайте свой сервер и начните общаться.
          </p>
        </div>

        {error && (
          <div className="mx-6 p-2.5 bg-red-500/10 border border-red-500/20 text-red-600 text-xs rounded text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-6 pt-2 space-y-4">
          <div className="space-y-2">
            <label className="text-xs font-bold text-[#4e5058] uppercase tracking-wider">Название сервера</label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Введите название сервера"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3 bg-[#e3e5e8] text-[#060607] rounded border-none focus:outline-none transition-colors text-base"
            />
            <p className="text-[10px] text-[#4e5058]">
              Создавая сервер, вы соглашаетесь с Правилами сообщества Minor.
            </p>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-[#f2f3f5] -mx-6 -mb-6 p-4 flex justify-between items-center">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-medium text-[#060607] hover:underline"
            >
              Назад
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-8 py-2.5 text-sm font-medium bg-[#5865f2] hover:bg-[#4752c4] text-white rounded transition-colors disabled:opacity-50"
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
