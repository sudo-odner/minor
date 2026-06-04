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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 font-sans backdrop-blur-[2px]">
      <div className="bg-white w-full max-w-md rounded-2xl overflow-hidden shadow-2xl animate-fade-in border border-transparent transition-all duration-200">
        
        {/* Шапка */}
        <div className="p-8 text-center">
          <h3 className="text-3xl font-bold text-gray-800 transition-colors">Создайте свой сервер</h3>
          <p className="text-gray-500 text-sm mt-3 leading-relaxed transition-colors">
            Ваш сервер — это место, где вы общаетесь с друзьями. Создайте свое сообщество прямо сейчас.
          </p>
        </div>

        {error && (
          <div className="mx-8 p-3 bg-red-50 border border-red-100 text-red-600 text-xs rounded-xl text-center transition-colors">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-8 pt-4 space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest px-1">Название сервера</label>
            <input
              type="text"
              required
              maxLength={100}
              placeholder="Введите название сервера"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3.5 bg-gray-50 text-gray-800 rounded-xl border border-gray-200 focus:outline-none focus:border-brand-blue focus:ring-1 focus:ring-brand-blue/10 transition-all text-base placeholder-gray-400 shadow-inner"
            />
            <p className="text-[10px] text-gray-400 text-center mt-2 px-4 transition-colors">
              Создавая сервер, вы соглашаетесь с Правилами сообщества Minor.
            </p>
          </div>

          {/* Подвал с кнопками */}
          <div className="bg-gray-50 -mx-8 -mb-8 p-6 flex justify-between items-center transition-colors duration-200 mt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 text-sm font-bold text-gray-500 hover:text-gray-800 hover:underline transition-colors"
            >
              Назад
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim()}
              className="px-10 py-3 text-sm font-bold bg-brand-blue hover:bg-brand-blue-dark text-white rounded-xl transition-all disabled:opacity-50 shadow-md active:scale-95"
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
