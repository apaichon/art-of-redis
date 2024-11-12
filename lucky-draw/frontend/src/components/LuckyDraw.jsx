import React, { useState, useEffect, useRef } from 'react';
import { Trophy } from 'lucide-react';

const WEBSOCKET_URL = `ws://${window.location.hostname}:9004/ws`;
const API_BASE_URL = `http://${window.location.hostname}:9004/api`;

const SlotNumber = ({ isSpinning, finalNumber, index }) => {
  const [numbers, setNumbers] = useState([0,1,2,3,4,5,6,7,8,9]);
  const containerRef = useRef(null);
  const animationRef = useRef(null);
  
  useEffect(() => {
    if (isSpinning) {
      startSpinning();
    } else if (finalNumber !== undefined) {
      stopSpinning(parseInt(finalNumber));
    }
    
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isSpinning, finalNumber]);

  const startSpinning = () => {
    if (containerRef.current) {
      containerRef.current.style.transition = 'none';
      containerRef.current.style.transform = 'translateY(0)';
      
      let start = null;
      let position = 0;
      
      const animate = (timestamp) => {
        if (!start) start = timestamp;
        const progress = timestamp - start;
        
        position = (progress / 5) % (numbers.length * 60);
        containerRef.current.style.transform = `translateY(-${position}px)`;
        
        if (isSpinning) {
          animationRef.current = requestAnimationFrame(animate);
        }
      };
      
      animationRef.current = requestAnimationFrame(animate);
    }
  };

  const stopSpinning = (finalNum) => {
    if (containerRef.current) {
      // Calculate final position based on the winning number
      const finalPosition = finalNum * 60; // 60px is the height of each number
      // Add extra rotations for a more dramatic effect
      const totalDistance = (numbers.length * 60 * 3) + finalPosition;
      
      containerRef.current.style.transition = `transform ${2 + index * 0.5}s cubic-bezier(0.23, 1, 0.32, 1)`;
      containerRef.current.style.transform = `translateY(-${totalDistance}px)`;
    }
  };

  return (
    <div className="w-16 h-16 bg-gray-900 rounded-lg relative overflow-hidden">
      <div
        ref={containerRef}
        className="absolute left-0 right-0 flex flex-col items-center"
        style={{
          willChange: 'transform',
        }}
      >
        {/* Repeat the numbers multiple times for continuous scrolling effect */}
        {[...numbers, ...numbers, ...numbers, ...numbers].map((num, idx) => (
          <div
            key={idx}
            className="w-16 h-[60px] flex items-center justify-center"
          >
            <span className="text-3xl font-bold text-white font-mono">
              {num}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

const LuckyDraw = () => {
  const [draw, setDraw] = useState(null);
  const [error, setError] = useState(null);
  const wsRef = useRef(null);
  const [isSpinning, setIsSpinning] = useState(false);
  const [winningNumbers, setWinningNumbers] = useState(['0', '0', '0', '0', '0', '0']);
  const [drawHistory, setDrawHistory] = useState([]);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    connectWebSocket();
    return () => cleanupWebSocket();
  }, []);

  const cleanupWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  };

  const connectWebSocket = () => {
    cleanupWebSocket();
    
    const ws = new WebSocket(WEBSOCKET_URL);
    
    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
    };
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === 'draw_update') {
          handleDrawUpdate(message.data);
        }
      } catch (err) {
        console.error('WebSocket message error:', err);
        setError('Failed to process server response');
      }
    };
    
    ws.onerror = () => {
      setError('Failed to connect');
      setIsConnected(false);
    };
    
    ws.onclose = () => {
      setIsConnected(false);
      setTimeout(connectWebSocket, 3000);
    };
    
    wsRef.current = ws;
  };

  const handleDrawUpdate = (drawData) => {
    setDraw(drawData);
    
    if (drawData.status === 'completed') {
      const finalNumbers = drawData.winningNumber.toString().padStart(6, '0').split('');
      setWinningNumbers(finalNumbers);
      
      // Allow time for the last number to finish spinning
      setTimeout(() => {
        setIsSpinning(false);
        setDrawHistory(prev => [drawData, ...prev].slice(0, 5));
      }, 5000); // Match the longest animation duration
    }
  };

  const startDraw = async () => {
    try {
      setError(null);
      if (!isConnected) {
        setError('Failed to connect. Please try again.');
        return;
      }

      setIsSpinning(true);
      setWinningNumbers(['0', '0', '0', '0', '0', '0']); // Reset numbers
      
      const response = await fetch(`${API_BASE_URL}/draw/start`, {
        method: 'POST',
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch');
      }
      
      const drawData = await response.json();
      console.log('drawData', drawData);
      setDraw(drawData);
    } catch (err) {
      setError('Failed to fetch');
      setIsSpinning(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#4338ca] flex items-center justify-center p-4">
      <div className="bg-white rounded-3xl shadow-2xl p-8 max-w-xl w-full">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-12 text-gray-900 flex items-center justify-center gap-2">
            Lucky Draw <Trophy className="text-yellow-500" size={32} />
          </h1>

          {/* Number Display */}
          <div className="flex justify-center gap-2 mb-8">
            {winningNumbers.map((number, index) => (
              <SlotNumber
                key={index}
                isSpinning={isSpinning}
                finalNumber={number}
                index={index}
              />
            ))}
          </div>

          {/* Error Display */}
          {error && (
            <div className="mb-6 text-red-500 bg-red-50 rounded-lg p-3">
              {error}
            </div>
          )}

          {/* Result Display */}
          {draw?.status === 'completed' && !isSpinning && (
            <div className="mb-6 text-green-700 bg-green-50 rounded-lg p-4">
              <p className="font-semibold">Draw Complete!</p>
              <p>Winning Number: {draw.winningNumber}</p>
            </div>
          )}

          {/* Draw Button */}
          <button
            onClick={startDraw}
            disabled={isSpinning}
            className="w-full bg-[#4ade80] text-white py-4 px-6 rounded-lg font-semibold text-lg
                     hover:bg-[#22c55e] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSpinning ? 'Drawing...' : 'Start New Draw'}
          </button>

          {/* Draw History */}
          {drawHistory.length > 0 && (
            <div className="mt-8">
              <h2 className="text-xl font-semibold text-gray-800 mb-4">Recent Draws</h2>
              <div className="space-y-2">
                {drawHistory.map((historyDraw, index) => (
                  <div 
                    key={index}
                    className="bg-gray-50 p-3 rounded-lg flex justify-between items-center"
                  >
                    <span className="text-gray-500 font-mono text-sm">
                      {historyDraw.id}
                    </span>
                    <span className="font-mono font-medium">
                      {historyDraw.winningNumber}
                    </span>
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