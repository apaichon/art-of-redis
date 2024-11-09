import React, { useState, useEffect, useRef } from 'react';
import { motion, animate } from 'framer-motion';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Trophy, RotateCw, Gift } from 'lucide-react';

const LuckyDraw = () => {
  const [isSpinning, setIsSpinning] = useState(false);
  const [currentDraw, setCurrentDraw] = useState(null);
  const [winningNumber, setWinningNumber] = useState(null);
  const [winners, setWinners] = useState([]);
  const wsRef = useRef(null);
  const wheelRef = useRef(null);

  useEffect(() => {
    connectWebSocket();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    wsRef.current = new WebSocket('ws://localhost:8080/ws');
    wsRef.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      if (message.type === 'draw_update') {
        handleDrawUpdate(message.data);
      } else if (message.type === 'winner_announcement') {
        handleWinnerAnnouncement(message.data);
      }
    };

    wsRef.current.onclose = () => {
      setTimeout(connectWebSocket, 5000);
    };
  };

  const handleDrawUpdate = (draw) => {
    setCurrentDraw(draw);
    if (draw.status === 'spinning') {
      setIsSpinning(true);
      spinWheel();
    } else if (draw.status === 'completed') {
      setIsSpinning(false);
      setWinningNumber(draw.winningNumber);
      showWinningNumber(draw.winningNumber);
    }
  };

  const handleWinnerAnnouncement = (winner) => {
    setWinners(prev => [winner, ...prev].slice(0, 10));
  };

  const startDraw = async () => {
    try {
      const response = await fetch('/api/draw/start', {
        method: 'POST'
      });
      if (!response.ok) throw new Error('Failed to start draw');
    } catch (error) {
      console.error('Error starting draw:', error);
    }
  };

  const spinWheel = () => {
    if (wheelRef.current) {
      animate(wheelRef.current, 
        { rotate: 360 * 5 }, 
        { duration: 5, ease: "easeOut" }
      );
    }
  };

  const showWinningNumber = (number) => {
    // Animate each digit appearing
    const digits = number.split('');
    digits.forEach((digit, index) => {
      setTimeout(() => {
        document.getElementById(`digit-${index}`).textContent = digit;
      }, index * 200);
    });
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto">
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="text-center text-3xl">Lucky Draw Wheel</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col items-center">
              {/* Wheel */}
              <motion.div
                ref={wheelRef}
                className="w-64 h-64 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 mb-8"
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 0 20px rgba(0,0,0,0.2)'
                }}
              >
                <div className="text-white text-4xl font-bold">
                  {isSpinning ? (
                    <RotateCw className="w-16 h-16 animate-spin" />
                  ) : (
                    <div className="grid grid-cols-6 gap-1">
                      {Array(6).fill(0).map((_, i) => (
                        <div
                          key={i}
                          id={`digit-${i}`}
                          className="w-8 h-12 bg-white bg-opacity-20 rounded flex items-center justify-center"
                        >
                          0
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </motion.div>

              {/* Controls */}
              <Button
                onClick={startDraw}
                disabled={isSpinning}
                className="px-8 py-4 text-lg"
              >
                {isSpinning ? 'Spinning...' : 'Start Draw'}
              </Button>

              {/* Winning Number Display */}
              {winningNumber && (
                <div className="mt-8 text-center">
                  <h3 className="text-xl font-semibold mb-2">Winning Number</h3>
                  <div className="text-4xl font-bold text-green-600">
                    {winningNumber}
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Recent Winners */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <Trophy className="w-6 h-6 mr-2" />
              Recent Winners
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {winners.map((winner, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-4 bg-white rounded-lg shadow"
                >
                  <div>
                    <div className="font-semibold">Number: {winner.number}</div>
                    <div className="text-sm text-gray-600">
                      {new Date(winner.claimedAt).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex items-center">
                    <Gift className="w-5 h-5 mr-2" />
                    <span className="font-medium">{winner.prize}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default LuckyDraw;
