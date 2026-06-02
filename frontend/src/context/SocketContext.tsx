import React, { createContext, useContext, useEffect, useState } from 'react';
import { useAuth } from './AuthContext';

interface SocketContextType {
  socket: WebSocket | null;
  isConnected: boolean;
}

const SocketContext = createContext<SocketContextType | undefined>(undefined);

export const SocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { accessToken, isAuthenticated } = useAuth();
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      if (socket) socket.close();
      return;
    }

    // Подключаемся к нашему Gateway Service через Traefik
    // Передаем токен в Query параметре для Handshake
    const ws = new WebSocket(`ws://localhost/gateway?token=${accessToken}`);

    ws.onopen = () => {
      console.log('WebSocket Connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('WS Message received:', data);
      // Тут будет логика обработки событий chat.message.created и т.д.
    };

    ws.onclose = () => {
      console.log('WebSocket Disconnected');
      setIsConnected(false);
    };

    setSocket(ws);

    return () => ws.close();
  }, [isAuthenticated, accessToken]);

  return (
    <SocketContext.Provider value={{ socket, isConnected }}>
      {children}
    </SocketContext.Provider>
  );
};

export const useSocket = () => {
  const context = useContext(SocketContext);
  if (!context) throw new Error('useSocket must be used within SocketProvider');
  return context;
};