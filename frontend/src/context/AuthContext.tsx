import React, { createContext, useContext, useEffect, useState } from "react";
import { api } from "../api/axios";

interface User {
    id: string;
    username: string;
    email: string;
}

interface AuthContextType {
    user: User | null;
    accessToken: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (data: { user: User; access_token: string }) => void;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);
let isRefreshing = false;

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
    children,
}) => {
    const [user, setUser] = useState<User | null>(null);
    const [accessToken, setAccessToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Функция для первичной проверки или обновления сессии
    const checkAuth = async () => {
        if (isRefreshing) return;
        isRefreshing = true;
        try {
            const response = await api.post("/api/v1/auth/refresh");
            const { access_token, user } = response.data;

            // ОБЯЗАТЕЛЬНО сохраняем новый токен, чтобы перехватчик Axios его увидел
            localStorage.setItem("accessToken", access_token);

            setAccessToken(access_token);
            setUser(user);
        } catch (e) {
            localStorage.removeItem("accessToken"); // Чистим при ошибке
            setAccessToken(null);
            setUser(null);
        } finally {
            setIsLoading(false); // Только теперь выключаем скелетон/загрузку
        }
    };

    useEffect(() => {
        checkAuth();
    }, []);

    const login = (data: { user: User; access_token: string }) => {
        setUser(data.user);
        setAccessToken(data.access_token);
        // Также можно сохранить в localStorage для надежности, если нужно
        localStorage.setItem("isLoggedIn", "true");
    };

    const logout = async () => {
        try {
            await api.post("/api/v1/auth/logout");
        } finally {
            setUser(null);
            setAccessToken(null);
            localStorage.removeItem("isLoggedIn");
            window.location.href = "/login";
        }
    };

    return (
        <AuthContext.Provider
            value={{
                user,
                accessToken,
                isAuthenticated: !!accessToken,
                isLoading,
                login,
                logout,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) throw new Error("useAuth must be used within AuthProvider");
    return context;
};
