// src/components/LuckyDraw.jsx
import React, { useState, useEffect, useRef } from 'react';
import { Trophy, RotateCw } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';

const WEBSOCKET_URL = `ws://${window.location.hostname}:9004/ws`;
const API_BASE_URL = `http://${window.location.hostname}:9004/api`;

const LuckyDraw = () => {
  const [draw, setDraw] = useState(null);
  const [error, setError] = useState(null);
  const wsRef = useRef(null);
  const numbersRef = useRef([]);
  const [animatingNumbers, setAnimatingNumbers] = useState(['0', '0', '0', '0', '0', '0']);
  const [drawHistory, setDrawHistory] = useState([]);

  useEffect(() => {
    connectWebSocket();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    const ws = new WebSocket(WEBSOCKET_URL);
    
    ws.onopen = () => {
      console.log('WebSocket Connected');
    };
    
    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'draw_update') {
        handleDrawUpdate(message.data);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket Error:', error);
      setError('Connection error. Please try again.');
    };
    
    ws.onclose = () => {
      console.log('WebSocket Disconnected');
      // Attempt to reconnect after 3 seconds
      setTimeout(connectWebSocket, 3000);
    };
    
    wsRef.current = ws;
  };

  const handleDrawUpdate = (drawData) => {
    setDraw(drawData);
    
    if (drawData.status === 'spinning') {
      startNumberAnimation();
    } else if (drawData.status === 'completed') {
      stopNumberAnimation(drawData.winningNumber);
      setDrawHistory(prev => [drawData, ...prev].slice(0, 5));
      playWinSound();
    }
  };

  const startNumberAnimation = () => {
    numbersRef.current = animatingNumbers.map(() => {
      return setInterval(() => {
        setAnimatingNumbers(prev => 
          prev.map(() => Math.floor(Math.random() * 10).toString())
        );
      }, 100);
    });
    playSpinSound();
  };

  const stopNumberAnimation = (winningNumber) => {
    const digits = winningNumber.toString().padStart(6, '0').split('');
    
    numbersRef.current.forEach((interval, index) => {
      setTimeout(() => {
        clearInterval(interval);
        setAnimatingNumbers(prev => {
          const newNumbers = [...prev];
          newNumbers[index] = digits[index];
          return newNumbers;
        });
      }, index * 200);
    });
  };

  const startDraw = async () => {
    try {
      setError(null);
      const response = await fetch(`${API_BASE_URL}/draw/start`, {
        method: 'POST',
      });
      
      if (!response.ok) {
        throw new Error('Failed to start draw');
      }
      
      const drawData = await response.json();
      handleDrawUpdate(drawData);
    } catch (err) {
      setError(err.message);
    }
  };

  const playSpinSound = () => {
    const audio = new Audio('/sounds/spin.mp3');
    audio.play().catch(() => {});
  };

  const playWinSound = () => {
    const audio = new Audio('/sounds/win.mp3');
    audio.play().catch(() => {});
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-600 to-purple-700 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl p-8 max-w-xl w-full">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-8 text-gray-800 flex items-center justify-center gap-2">
            Lucky Draw <Trophy className="text-yellow-500" />
          </h1>

          {/* Number Display */}
          <div className="mb-8 relative">
            <div className="flex justify-center gap-2 mb-6">
              {animatingNumbers.map((number, index) => (
                <div
                  key={index}
                  className="w-16 h-20 bg-gradient-to-b from-gray-800 to-gray-900 rounded-lg flex items-center justify-center 
                           shadow-lg transform transition-all duration-200 hover:scale-105"
                >
                  <span className="text-4xl font-bold text-white font-mono">
                    {number}
                  </span>
                </div>
              ))}
            </div>

            {draw?.status === 'spinning' && (
              <div className="flex items-center justify-center gap-2 text-purple-600">
                <RotateCw className="animate-spin" />
                <span>Drawing in progress...</span>
              </div>
            )}
          </div>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Draw Button */}
          {draw?.status === 'completed' ? (
            <div className="space-y-4">
              <div className="p-4 bg-green-50 rounded-lg">
                <p className="text-green-800 font-medium">Draw Complete!</p>
                <p className="text-green-600">
                  Winning Number: {draw.winningNumber}
                </p>
              </div>
              <button
                onClick={startDraw}
                className="w-full bg-green-500 text-white py-3 px-6 rounded-lg font-medium 
                         hover:bg-green-600 transition-colors"
              >
                Start New Draw
              </button>
            </div>
          ) : (
            <button
              onClick={startDraw}
              disabled={draw?.status === 'spinning'}
              className={`w-full bg-gradient-to-r from-blue-500 to-purple-500 text-white py-3 px-6 
                       rounded-lg font-medium transition-all duration-300
                       ${draw?.status === 'spinning' 
                         ? 'opacity-50 cursor-not-allowed' 
                         : 'hover:from-blue-600 hover:to-purple-600'}`}
            >
              {draw?.status === 'spinning' ? 'Drawing...' : 'Start Draw'}
            </button>
          )}

          {/* Draw History */}
          {drawHistory.length > 0 && (
            <div className="mt-8">
              <h2 className="text-lg font-semibold text-gray-700 mb-3">Recent Draws</h2>
              <div className="space-y-2">
                {drawHistory.map((historyDraw, index) => (
                  <div 
                    key={index}
                    className="bg-gray-50 p-2 rounded-lg flex justify-between items-center"
                  >
                    <span className="text-gray-600">#{historyDraw.id}</span>
                    <span className="font-mono font-medium">{historyDraw.winningNumber}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LuckyDraw;