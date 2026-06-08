import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/axios';

const ForgotPasswordPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');
        setLoading(true);
        try {
            await api.post('/auth/forgot-password', { email });
            setMessage('Инструкции по восстановлению пароля отправлены на ваш email.');
        } catch (err: any) {
            const errMsg = err.response?.data?.error?.message || err.response?.data?.error || 'Произошла ошибка. Попробуйте позже.';
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
                            <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3y-3.5L15.5 7.5z"/>
                        </svg>
                    </div>
                    <h1 className="text-2xl font-bold text-gray-800 tracking-tight transition-colors">Восстановление пароля</h1>
                    <p className="text-gray-500 mt-2 text-sm transition-colors">Введите ваш email, чтобы получить инструкции</p>
                </div>

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
                        <label className="block text-xs font-bold text-gray-500 uppercase mb-1 ml-1 transition-colors">Email</label>
                        <input 
                            type="email" 
                            placeholder="email@example.com" 
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 rounded-xl bg-gray-50 border border-gray-200 text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[#002FA7]/20 focus:border-[#002FA7] transition-all"
                            required
                        />
                    </div>
                    
                    <button 
                        type="submit" 
                        disabled={loading}
                        className="w-full bg-[#002FA7] hover:bg-[#001f7a] text-white font-bold py-3.5 rounded-xl shadow-lg active:scale-[0.98] transition-all disabled:opacity-70 mt-2"
                    >
                        {loading ? 'Отправка...' : 'Сбросить пароль'}
                    </button>
                </form>

                <div className="mt-8 text-center text-sm text-gray-500 font-medium transition-colors">
                    Вспомнили пароль? 
                    <Link to="/login" className="ml-1 text-[#002FA7] hover:underline">
                        Войти
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default ForgotPasswordPage;
