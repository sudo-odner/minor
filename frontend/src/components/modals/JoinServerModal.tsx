import React, { useState } from "react";
import { joinServer } from "../../api/members";

interface JoinServerModalProps {
    isOpen: boolean;
    onClose: () => void;
    onJoined: (newServer: any) => void;
}

const JoinServerModal: React.FC<JoinServerModalProps> = ({
    isOpen,
    onClose,
    onJoined,
}) => {
    const [serverId, setServerId] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    if (!isOpen) return null;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!serverId.trim()) return;

        setLoading(true);
        setError("");

        try {
            const result = await joinServer(serverId.trim());
            onJoined(result); 
            onClose();
        } catch (err: any) {
            if (err.response?.status === 409) {
                setError("Вы уже являетесь участником этого сервера");
            } else {
                setError(
                    err.response?.data?.message ||
                        "Сервер не найден или произошла ошибка",
                );
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 font-sans backdrop-blur-[2px]">
            <div className="bg-white w-full max-w-md rounded-2xl overflow-hidden shadow-2xl border border-transparent p-8 transition-all duration-200">
                <h3 className="text-3xl font-bold text-gray-800 mb-3 transition-colors text-center">
                    Присоединиться
                </h3>
                <p className="text-gray-500 text-sm mb-6 text-center leading-relaxed transition-colors">
                    Введите ID сервера (UUID), чтобы стать частью нового сообщества.
                </p>

                {error && (
                    <div className="mb-6 p-3 bg-red-50 border border-red-100 text-red-600 text-xs rounded-xl text-center transition-colors">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                    <input
                        type="text"
                        required
                        placeholder="019e84d7-b837-7d9b-..."
                        value={serverId}
                        onChange={(e) => setServerId(e.target.value)}
                        className="w-full px-4 py-3.5 bg-gray-50 text-gray-800 rounded-xl border border-gray-200 focus:outline-none focus:border-brand-blue focus:ring-1 focus:ring-brand-blue/10 transition-all text-sm placeholder-gray-400"
                    />

                    <div className="flex justify-between items-center pt-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-bold text-gray-500 hover:text-gray-800 hover:underline transition-colors"
                        >
                            Отмена
                        </button>
                        <button
                            type="submit"
                            disabled={loading || !serverId.trim()}
                            className="px-10 py-3 bg-brand-blue hover:bg-brand-blue-dark text-white rounded-xl text-sm font-bold disabled:opacity-50 shadow-md active:scale-95 transition-all"
                        >
                            {loading ? "..." : "Войти"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default JoinServerModal;
