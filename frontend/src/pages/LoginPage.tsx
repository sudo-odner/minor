import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'; // Импортируем навигатор
import { api } from '../api/axios';
import { useAuth } from '../context/AuthContext'; // Импортируем наш хук

const LoginPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    
    const { login } = useAuth(); // Получаем функцию входа из контекста
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const response = await api.post('/api/v1/auth/login', {
                email,
                password,
            });

            // 1. Сохраняем токен в localStorage (чтобы axios его подхватил)
            localStorage.setItem('accessToken', response.data.access_token);

            // 2. Обновляем состояние в AuthContext (это заставит App.tsx переключить роут)
            login(response.data); 

            // 3. Плавный переход на главную
            navigate('/');
        } catch (err: any) {
            setError(err.response?.data?.error || 'Неверный логин или пароль');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-[#e6f0ff] flex items-center justify-center min-h-screen p-4 font-sans">
            <div className="bg-white w-full max-w-md rounded-2xl shadow-xl p-8 sm:p-10">
                <div className="text-center mb-10">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#e6f0ff] text-[#002FA7] mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                    </div>
                    <h1 className="text-2xl font-bold text-gray-800">Добро пожаловать</h1>
                </div>

                {error && <div className="mb-4 text-red-500 text-center text-sm">{error}</div>}

                <form onSubmit={handleSubmit} className="space-y-6">
                    <input 
                        type="email" 
                        placeholder="Email" 
                        className="w-full px-4 py-3 rounded-xl border"
                        value={email}
                        onChange={e => setEmail(e.target.value)}
                        required
                    />
                    <input 
                        type="password" 
                        placeholder="Пароль" 
                        className="w-full px-4 py-3 rounded-xl border"
                        value={password}
                        onChange={e => setPassword(e.target.value)}
                        required
                    />
                    <button 
                        type="submit" 
                        disabled={loading}
                        className="w-full bg-[#002FA7] text-white py-3 rounded-xl hover:bg-[#0047E6] transition-all"
                    >
                        {loading ? 'Вход...' : 'Войти'}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default LoginPage;