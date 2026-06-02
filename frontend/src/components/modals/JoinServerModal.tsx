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
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 font-sans">
            <div className="bg-[#313338] w-full max-w-md rounded-lg overflow-hidden shadow-2xl border border-[#1e1f22] p-6">
                <h3 className="text-xl font-bold text-white mb-2">
                    Вступить по приглашению
                </h3>
                <p className="text-gray-400 text-xs mb-4">
                    Введите ID сервера (UUID), чтобы присоединиться к
                    сообществу.
                </p>

                {error && (
                    <div className="mb-4 p-2.5 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded text-center">
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
                        className="w-full px-4 py-3 bg-[#1e1f22] text-white rounded border border-[#1e1f22] focus:outline-none focus:border-[#5865f2] text-sm"
                    />

                    <div className="flex justify-end space-x-3 pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-white hover:underline"
                        >
                            Отмена
                        </button>
                        <button
                            type="submit"
                            disabled={loading || !serverId.trim()}
                            className="px-6 py-2 bg-[#5865f2] hover:bg-[#4752c4] text-white rounded text-sm font-medium disabled:opacity-50"
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
