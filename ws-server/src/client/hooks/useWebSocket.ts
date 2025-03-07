import { useState, useEffect, useCallback, useRef } from 'react';

interface WebSocketMessage {
  type: string;
  [key: string]: any;
}

interface UseWebSocketOptions {
  onOpen?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (event: Event) => void;
  reconnectInterval?: number;
  reconnectAttempts?: number;
}

interface UseWebSocketReturn {
  isConnected: boolean;
  messages: WebSocketMessage[];
  sendMessage: (message: WebSocketMessage) => void;
  clearMessages: () => void;
}

const useWebSocket = (
  url: string,
  options: UseWebSocketOptions = {}
): UseWebSocketReturn => {
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<WebSocketMessage[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCountRef = useRef(0);
  const reconnectTimerRef = useRef<number | null>(null);

  const {
    onOpen,
    onClose,
    onError,
    reconnectInterval = 3000,
    reconnectAttempts = 5,
  } = options;

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    wsRef.current = new WebSocket(url);

    wsRef.current.onopen = (event) => {
      setIsConnected(true);
      reconnectCountRef.current = 0;
      if (onOpen) onOpen(event);
    };

    wsRef.current.onclose = (event) => {
      setIsConnected(false);
      if (onClose) onClose(event);

      // Attempt to reconnect
      if (reconnectCountRef.current < reconnectAttempts) {
        reconnectCountRef.current += 1;
        reconnectTimerRef.current = window.setTimeout(() => {
          connect();
        }, reconnectInterval);
      }
    };

    wsRef.current.onerror = (event) => {
      if (onError) onError(event);
    };

    wsRef.current.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setMessages((prevMessages) => [...prevMessages, data]);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };
  }, [url, onOpen, onClose, onError, reconnectAttempts, reconnectInterval]);

  const sendMessage = useCallback(
    (message: WebSocketMessage) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify(message));
      } else {
        console.error('WebSocket is not connected');
      }
    },
    []
  );

  useEffect(() => {
    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
    };
  }, [connect]);

  return {
    isConnected,
    messages,
    sendMessage,
    clearMessages,
  };
};

export default useWebSocket; 