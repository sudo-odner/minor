import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';

// Контексты
import { AuthProvider, useAuth } from './context/AuthContext';
import { SocketProvider } from './context/SocketContext';

// Страницы и Лейауты
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import MainLayout from './layouts/MainLayout';

import { ThemeProvider, useTheme } from './context/ThemeContext';

/**
 * ProtectedRoute - Охранник роутов.
 * 1. Пока isLoading = true (идет запрос /refresh), показываем спиннер.
 * 2. Если не авторизован - редирект на /login.
 * 3. Если авторизован - рендерим дочерние компоненты.
 */
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, isLoading } = useAuth();
  const { theme } = useTheme();

  if (isLoading) {
    return (
      <div className="h-screen w-full bg-white dark:bg-[#313338] flex items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[#5865f2]"></div>
          <span className="text-gray-400 text-sm font-medium animate-pulse">Загрузка Minor...</span>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

/**
 * AppContent - Внутренняя часть приложения, где доступен роутинг.
 * Вынесена отдельно, чтобы можно было использовать useAuth внутри.
 */
const AppContent: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* Публичные маршруты */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* Защищенные маршруты (Вся основная логика мессенджера) */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MainLayout>
                {/* 
                  Этот контент отображается в центре MainLayout (в блоке {children}), 
                  если не выбран конкретный канал. 
                */}
                <div className="flex-1 flex flex-col items-center justify-center text-center p-8 select-none bg-white dark:bg-[#313338] transition-colors">
                  <div className="w-24 h-24 bg-[#ebedef] dark:bg-[#35363c] rounded-full flex items-center justify-center mb-6 shadow-inner transition-colors">
                    <span className="text-5xl opacity-50">👋</span>
                  </div>
                  <h2 className="text-2xl font-bold text-[#060607] dark:text-white mb-2 transition-colors">
                    Добро пожаловать в Minor!
                  </h2>
                  <p className="text-[#4f5660] dark:text-gray-400 max-w-sm leading-relaxed transition-colors">
                    Выберите сервер слева или вступите в сообщество по ID, чтобы начать общение с друзьями.
                  </p>
                </div>
              </MainLayout>
            </ProtectedRoute>
          }
        />

        {/* Редирект со всех остальных путей на главную */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

/**
 * App - Корневой компонент.
 * Порядок провайдеров важен: Сокету нужен токен из AuthContext.
 */
function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <SocketProvider>
          <AppContent />
        </SocketProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;