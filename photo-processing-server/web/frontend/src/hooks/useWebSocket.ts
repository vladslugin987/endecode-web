import { useEffect, useRef, useCallback } from 'react';

export interface WebSocketMessage {
  type: 'log' | 'progress' | 'status' | 'complete' | 'error';
  data: {
    message?: string;
    progress?: number;
    status?: string;
    error?: string;
    result?: any;
  };
}

export interface ClientMessage {
  type: 'subscribe' | 'unsubscribe';
  jobId?: string;
}

interface UseWebSocketProps {
  onMessage: (message: WebSocketMessage) => void;
  url?: string;
  onOpen?: () => void;
  onClose?: (code?: number, reason?: string) => void;
}

export const useWebSocket = ({ onMessage, url = `${window.location.origin.replace('http', 'ws')}/ws`, onOpen, onClose }: UseWebSocketProps) => {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;

  const connect = useCallback(() => {
    // Avoid creating a new connection if one is already open
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      return;
    }
    try {
      const ws = new WebSocket(url);
      
      ws.onopen = () => {
        console.log('WebSocket connected');
        reconnectAttempts.current = 0;
        wsRef.current = ws;
        if (onOpen) onOpen();
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          onMessage(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        wsRef.current = null;
        if (onClose) onClose(event.code, event.reason);
        
        // Attempt to reconnect
        if (reconnectAttempts.current < maxReconnectAttempts) {
          reconnectAttempts.current++;
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000);
          console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts.current}/${maxReconnectAttempts})`);
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        } else {
          console.error('Max reconnection attempts reached');
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  }, [url, onMessage]);

  const sendMessage = useCallback((message: ClientMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected, message not sent:', message);
    }
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    sendMessage,
    disconnect,
    isConnected: wsRef.current?.readyState === WebSocket.OPEN
  };
};