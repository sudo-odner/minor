import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../api/axios';

const RegisterPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);
        try {
            await api.post('/auth/register', { email, username, password });
            navigate('/login');
        } catch (err: any) {
            const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Ошибка при регистрации';
            setError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-[#e6f0ff] dark:bg-[#1e1f22] flex items-center justify-center min-h-screen p-4 font-sans transition-colors duration-200">
            <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-2xl shadow-xl p-8 sm:p-10 transition-all duration-300 border border-transparent dark:border-[#1e1f22]">
                
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#e6f0ff] dark:bg-[#404249] text-[#002FA7] dark:text-[#5865f2] mb-4 transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/>
                        </svg>
                    </div>
                    <h1 className="text-2xl font-bold text-gray-800 dark:text-white tracking-tight transition-colors">Создать аккаунт</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-sm transition-colors">Присоединяйтесь к сообществу Minor</p>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-50 dark:bg-red-500/10 text-red-600 dark:text-red-400 text-xs rounded-lg border border-red-100 dark:border-red-500/20 text-center font-medium transition-colors">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-1 ml-1 transition-colors">Email</label>
                        <input 
                            type="email" 
                            placeholder="email@example.com" 
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 dark:bg-[#1e1f22] border border-gray-200 dark:border-transparent text-gray-800 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 dark:focus:ring-[#5865f2]/20 focus:border-[#002FA7] dark:focus:border-[#5865f2] transition-all"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-1 ml-1 transition-colors">Имя пользователя</label>
                        <input 
                            type="text" 
                            placeholder="username" 
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 dark:bg-[#1e1f22] border border-gray-200 dark:border-transparent text-gray-800 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 dark:focus:ring-[#5865f2]/20 focus:border-[#002FA7] dark:focus:border-[#5865f2] transition-all"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-1 ml-1 transition-colors">Пароль</label>
                        <input 
                            type="password" 
                            placeholder="••••••••" 
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 dark:bg-[#1e1f22] border border-gray-200 dark:border-transparent text-gray-800 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 dark:focus:ring-[#5865f2]/20 focus:border-[#002FA7] dark:focus:border-[#5865f2] transition-all"
                            required
                        />
                    </div>
                    
                    <button 
                        type="submit" 
                        disabled={loading}
                        className="w-full bg-[#002FA7] dark:bg-[#5865f2] hover:bg-[#001f7a] dark:hover:bg-[#4752c4] text-white font-bold py-3.5 rounded-xl shadow-lg active:scale-[0.98] transition-all disabled:opacity-70 mt-2"
                    >
                        {loading ? 'Создание...' : 'Зарегистрироваться'}
                    </button>
                </form>

                <div className="mt-8 text-center text-sm text-gray-500 dark:text-gray-400 font-medium transition-colors">
                    Уже есть аккаунт? 
                    <Link to="/login" className="ml-1 text-[#002FA7] dark:text-[#5865f2] hover:underline">
                        Войти
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default RegisterPage;