import { useState, useEffect, useRef } from 'react';
import { WS_URL } from '../config/env';

export const useWebSocket = () => {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [isConnecting, setIsConnecting] = useState(true);
  const ws = useRef(null);
  const reconnectTimeout = useRef(null);

  const connect = () => {
    try {
      ws.current = new WebSocket(WS_URL);

      ws.current.onopen = () => {
        console.log('WebSocket Connected');
        setIsConnecting(false);
        setError(null);
      };

      ws.current.onmessage = (event) => {
        try {
          const parsedData = JSON.parse(event.data);
          setData(parsedData);
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };

      ws.current.onclose = () => {
        console.log('WebSocket Disconnected');
        setIsConnecting(true);
        // Attempt to reconnect after 3 seconds
        reconnectTimeout.current = setTimeout(connect, 3000);
      };

      ws.current.onerror = (event) => {
        console.error('WebSocket Error:', event);
        setError('Failed to connect to server');
        ws.current?.close();
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to connect to server');
      setIsConnecting(true);
    }
  };

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
      }
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  return {
    data,
    error,
    isConnecting
  };
};