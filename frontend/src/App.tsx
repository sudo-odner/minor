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
      <div className="h-screen w-full bg-brand-bg flex items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-brand-blue"></div>
          <span className="text-brand-blue/60 text-sm font-bold tracking-widest uppercase animate-pulse">Загрузка Minor...</span>
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
                <div className="flex-1 flex flex-col items-center justify-center text-center p-8 select-none bg-white transition-colors">
                  <div className="w-32 h-32 bg-brand-blue-light/50 rounded-full flex items-center justify-center mb-8 shadow-inner transition-colors">
                    <span className="text-6xl opacity-70">👋</span>
                  </div>
                  <h2 className="text-3xl font-extrabold text-gray-800 mb-3 transition-colors tracking-tight">
                    Добро пожаловать в Minor!
                  </h2>
                  <p className="text-gray-500 max-w-sm leading-relaxed transition-colors font-medium">
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