import React, { useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { api } from '../api/axios';

const ResetPasswordPage: React.FC = () => {
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token');
    
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');

        if (password !== confirmPassword) {
            setError('Пароли не совпадают');
            return;
        }

        if (!token) {
            setError('Токен сброса пароля отсутствует или недействителен');
            return;
        }

        setLoading(true);
        try {
            await api.post('/auth/reset-password', { token, password });
            setMessage('Пароль успешно изменен. Теперь вы можете войти.');
        } catch (err: any) {
            const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Произошла ошибка при сбросе пароля.';
            setError(typeof errMsg === 'string' ? errMsg : JSON.stringify(errMsg));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-[#e6f0ff] flex items-center justify-center min-h-screen p-4 font-sans transition-colors duration-200">
            <div className="bg-white w-full max-w-md rounded-2xl shadow-xl p-8 sm:p-10 transition-all duration-300 border border-transparent">
                
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#e6f0ff] text-[#002FA7] mb-4 transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M8 11h8"/>
                        </svg>
                    </div>
                    <h1 className="text-2xl font-bold text-gray-800 tracking-tight transition-colors">Новый пароль</h1>
                    <p className="text-gray-500 mt-2 text-sm transition-colors">Придумайте надежный пароль для вашего аккаунта</p>
                </div>

                {!token && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 text-xs rounded-lg border border-red-100 text-center font-medium transition-colors">
                        Токен сброса пароля не найден. Пожалуйста, воспользуйтесь ссылкой из письма.
                    </div>
                )}

                {error && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 text-xs rounded-lg border border-red-100 text-center font-medium transition-colors">
                        {error}
                    </div>
                )}

                {message && (
                    <div className="mb-4 p-3 bg-green-50 text-green-600 text-xs rounded-lg border border-green-100 text-center font-medium transition-colors">
                        {message}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-xs font-bold text-gray-500 uppercase mb-1 ml-1 transition-colors">Новый пароль</label>
                        <input 
                            type="password" 
                            placeholder="••••••••" 
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 border border-gray-200 text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 focus:border-[#002FA7] transition-all"
                            required
                            disabled={!token || !!message}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-gray-500 uppercase mb-1 ml-1 transition-colors">Подтвердите пароль</label>
                        <input 
                            type="password" 
                            placeholder="••••••••" 
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 border border-gray-200 text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 focus:border-[#002FA7] transition-all"
                            required
                            disabled={!token || !!message}
                        />
                    </div>
                    
                    <button 
                        type="submit" 
                        disabled={loading || !token || !!message}
                        className="w-full bg-[#002FA7] hover:bg-[#001f7a] text-white font-bold py-3.5 rounded-xl shadow-lg active:scale-[0.98] transition-all disabled:opacity-70 mt-2"
                    >
                        {loading ? 'Сохранение...' : 'Сбросить пароль'}
                    </button>
                </form>

                <div className="mt-8 text-center text-sm text-gray-500 font-medium transition-colors">
                    <Link to="/login" className="text-[#002FA7] hover:underline">
                        Вернуться к входу
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default ResetPasswordPage;
