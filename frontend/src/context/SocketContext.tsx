import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useAuth } from './AuthContext';

interface SocketContextType {
  socket: WebSocket | null;
  isConnected: boolean;
}

const SocketContext = createContext<SocketContextType | undefined>(undefined);

// Global event target to dispatch WS messages independently of React render cycle
export const wsEventBus = new EventTarget();

export const SocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { accessToken, isAuthenticated } = useAuth();
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      if (wsRef.current) wsRef.current.close();
      setSocket(null);
      setIsConnected(false);
      return;
    }

    let destroyed = false;

    const connect = () => {
      if (destroyed) return;

      const ws = new WebSocket(`ws://localhost/gateway?token=${accessToken}`);
      wsRef.current = ws;

      ws.onopen = () => {
        if (destroyed) { ws.close(); return; }
        console.log('WebSocket Connected');
        setIsConnected(true);
        setSocket(ws);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('WS Message received:', data);
          // Dispatch to global event bus so all components receive it
          // regardless of React render timing
          wsEventBus.dispatchEvent(new CustomEvent('ws_message', { detail: data }));
        } catch {}
      };

      ws.onclose = () => {
        if (destroyed) return;
        console.log('WebSocket Disconnected, reconnecting in 2s...');
        setIsConnected(false);
        setSocket(null);
        reconnectTimer.current = setTimeout(connect, 2000);
      };

      ws.onerror = (err) => {
        console.warn('WebSocket error', err);
        ws.close();
      };
    };

    connect();

    return () => {
      destroyed = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (wsRef.current) wsRef.current.close();
    };
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