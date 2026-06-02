import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../api/axios';
import { useAuth } from '../context/AuthContext';

const LoginPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);
        try {
            const response = await api.post('/auth/login', { email, password });
            localStorage.setItem('accessToken', response.data.access_token);
            login(response.data); 
            navigate('/');
        } catch (err: any) {
            setError(err.response?.data?.error || 'Неверный логин или пароль');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-[#e6f0ff] flex items-center justify-center min-h-screen p-4 font-sans">
            <div className="bg-white w-full max-w-md rounded-2xl shadow-xl p-8 sm:p-10 transition-all duration-300">
                
                {/* Исправленный блок иконки */}
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#e6f0ff] text-[#002FA7] mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/>
                        </svg>
                    </div>
                    <h1 className="text-2xl font-bold text-gray-800 tracking-tight">Добро пожаловать</h1>
                    <p className="text-gray-500 mt-2 text-sm">Войдите в свой аккаунт Minor</p>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 text-xs rounded-lg border border-red-100 text-center font-medium">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-xs font-bold text-gray-500 uppercase mb-1 ml-1">Email</label>
                        <input 
                            type="email" 
                            placeholder="email@example.com" 
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 border border-gray-200 text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 focus:border-[#002FA7] transition-all"
                            required
                        />
                    </div>
                    <div>
                        <div className="flex justify-between items-center mb-1 ml-1">
                            <label className="block text-xs font-bold text-gray-500 uppercase">Пароль</label>
                            <a href="#" className="text-xs font-semibold text-[#002FA7] hover:underline">Забыли?</a>
                        </div>
                        <input 
                            type="password" 
                            placeholder="••••••••" 
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 border border-gray-200 text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 focus:border-[#002FA7] transition-all"
                            required
                        />
                    </div>
                    
                    <button 
                        type="submit" 
                        disabled={loading}
                        className="w-full bg-[#002FA7] hover:bg-[#001f7a] text-white font-bold py-3.5 rounded-xl shadow-lg active:scale-[0.98] transition-all disabled:opacity-70 mt-2"
                    >
                        {loading ? 'Вход...' : 'Войти'}
                    </button>
                </form>

                <div className="mt-8 text-center text-sm text-gray-500 font-medium">
                    Нет аккаунта? 
                    <Link to="/register" className="ml-1 text-[#002FA7] hover:underline">
                        Создать аккаунт
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;