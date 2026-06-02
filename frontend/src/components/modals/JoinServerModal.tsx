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
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans backdrop-blur-[2px]">
            <div className="bg-white dark:bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl border border-[#e3e5e8] dark:border-[#1e1f22] p-6 transition-colors duration-200">
                <h3 className="text-xl font-bold text-[#060607] dark:text-white mb-2 transition-colors">
                    Вступить по приглашению
                </h3>
                <p className="text-[#4f5660] dark:text-gray-400 text-xs mb-4 transition-colors">
                    Введите ID сервера (UUID), чтобы присоединиться к
                    сообществу.
                </p>

                {error && (
                    <div className="mb-4 p-2.5 bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded text-center transition-colors">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <input
                        type="text"
                        required
                        placeholder="019e84d7-b837-7d9b-..."
                        value={serverId}
                        onChange={(e) => setServerId(e.target.value)}
                        className="w-full px-4 py-3 bg-[#ebedef] dark:bg-[#1e1f22] text-[#060607] dark:text-white rounded border border-transparent focus:outline-none focus:border-[#5865f2] transition-all text-sm placeholder-[#4f5660] dark:placeholder-gray-500"
                    />

                    <div className="flex justify-end space-x-3 pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-[#060607] dark:text-white hover:underline transition-colors"
                        >
                            Отмена
                        </button>
                        <button
                            type="submit"
                            disabled={loading || !serverId.trim()}
                            className="px-6 py-2 bg-[#5865f2] hover:bg-[#4752c4] text-white rounded text-sm font-medium disabled:opacity-50 shadow-sm"
                        >
                            {loading ? "Вступление..." : "Присоединиться"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default JoinServerModal;
