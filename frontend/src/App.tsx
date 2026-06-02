import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { SocketProvider } from './context/SocketContext';

// Импорт страниц
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import MainLayout from './layouts/MainLayout';

// Компонент для защиты приватных маршрутов
// Если пользователь не залогинен, его перекинет на /login
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, isLoading } = useAuth();

  // Если мы еще проверяем куку — ничего не делаем, показываем пустой экран или лоадер
  if (isLoading) {
    return <div className="h-screen bg-[#313338]" />; 
  }

  // Только когда загрузка завершена, решаем: пускать или нет
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

const AppContent: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* Публичные маршруты */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* Защищенные маршруты (наш мессенджер) */}
        <Route 
          path="/" 
          element={
            <ProtectedRoute>
              <MainLayout>
                {/* Внутри MainLayout мы будем отображать сообщения */}
                <div className="flex flex-col items-center justify-center h-full text-gray-400">
                  <h2 className="text-2xl font-bold mb-2">Добро пожаловать в Minor!</h2>
                  <p>Выберите канал, чтобы начать общение.</p>
                </div>
              </MainLayout>
            </ProtectedRoute>
          } 
        />

        {/* Редирект со всех неизвестных путей на главную */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  );
};

function App() {
  return (
    // Важно: AuthProvider должен быть выше SocketProvider, 
    // так как сокету нужен токен из AuthContext
    <AuthProvider>
      <SocketProvider>
        <AppContent />
      </SocketProvider>
    </AuthProvider>
  );
}

export default App;