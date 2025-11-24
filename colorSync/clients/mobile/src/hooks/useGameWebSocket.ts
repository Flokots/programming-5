import { useState, useEffect, useRef, useCallback } from "react";
import type {
    WSMessage,
    GameState,
    GameStartPayload,
    RoundStartPayload,
    RoundResultPayload,
    GameOverPayload,
    ErrorPayload,
    ClickPayload,
    StroopColor
} from '../types/game';

export function useGameWebSocket(roomId: string, userId: string) {
    const [gameState, setGameState] = useState<GameState | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const hasConnectedRef = useRef(false);

    // ✅ Handler: Game Start
    const handleGameStart = useCallback((payload: GameStartPayload) => {
        console.log('🎮 Game starting with', payload.max_rounds, 'rounds');
        setGameState(prev => {
            if (!prev) return null;
            return {
                ...prev,
                status: 'playing',
                maxRounds: payload.max_rounds,
                currentRound: 0,
                myScore: 0,
                opponentScore: 0
            };
        });
    }, []);

    // ✅ Handler: Round Start
    const handleRoundStart = useCallback((payload: RoundStartPayload) => {
        console.log('🎨 Round', payload.round, ':', payload.word, '(', payload.color, ')');
        setGameState(prev => {
            if (!prev) return null;
            return {
                ...prev,
                status: 'playing',
                currentRound: payload.round,
                word: payload.word,
                color: payload.color
            };
        });
    }, []);

    // ✅ Handler: Round Result - Score tracking
    const handleRoundResult = useCallback((payload: RoundResultPayload) => {
        console.log('📊 Round result - Winner:', payload.winner);
        console.log('📊 Current user ID:', userId);

        setGameState(prev => {
            if (!prev) return prev;

            let newMyScore = prev.myScore;
            let newOpponentScore = prev.opponentScore;

            if (payload.winner === userId) {
                // I won this round
                newMyScore += 1;
                console.log('✅ I won! My score:', newMyScore);
            } else if (payload.winner !== null && payload.winner !== userId) {
                // Opponent won this round
                newOpponentScore += 1;
                console.log('❌ Opponent won! Their score:', newOpponentScore);
            } else {
                // No winner (timeout)
                console.log('⏱️ Round timed out - no winner');
            }

            return {
                ...prev,
                myScore: newMyScore,
                opponentScore: newOpponentScore
            };
        });
    }, [userId]);

    // ✅ Handler: Game Over
    const handleGameOver = useCallback((payload: GameOverPayload) => {
        console.log('🏁 Game over - Winner:', payload.winner);
        console.log('📊 Final stats:', payload.stats);
        
        setGameState(prev => {
            if (!prev) return null;
            return {
                ...prev,
                status: 'game_over',
                winner: payload.winner,
                myStats: payload.stats?.[userId] || null
            };
        });
    }, [userId]);

    // ✅ Handler: Error
    const handleError = useCallback((payload: ErrorPayload) => {
        console.error('❌ Error from server:', payload.message);
        setGameState(prev => {
            if (!prev) return null;
            return {
                ...prev,
                errorMessage: payload.message
            };
        });
    }, []);

    // WebSocket connection effect
    useEffect(() => {
        if (!roomId || !userId) {
            console.log('⏸️ Skipping WebSocket connection (no roomId or userId)');
            return;
        }

        if (hasConnectedRef.current || wsRef.current) {
            console.log('⚠️ Already connected or connecting, skipping reconnection');
            return;
        }

        console.log('🔌 Connecting WebSocket:', { roomId, userId });
        hasConnectedRef.current = true;

        // Initialize game state
        setGameState({
            status: 'connecting',
            currentRound: 0,
            maxRounds: 0,
            word: '',
            color: '',
            myScore: 0,
            opponentScore: 0,
            winner: null,
            myStats: null,
            errorMessage: null,
            userId,
            roomId
        });

        // ⚠️ IMPORTANT: For physical devices, change to ws://YOUR_IP:8003
        const ws = new WebSocket(
            `ws://192.168.30.152:8003/game/ws?room_id=${roomId}&user_id=${userId}`
        );

        ws.onopen = () => {
            console.log('✅ WebSocket connected');
            setIsConnected(true);
            wsRef.current = ws;
            setGameState(prev => prev ? { ...prev, status: 'waiting' } : null);
        };

        ws.onmessage = (event) => {
            try {
                const message: WSMessage = JSON.parse(event.data);
                console.log('📨 Received message:', message.type);

                // Route to appropriate handler
                switch (message.type) {
                    case 'GAME_START':
                        handleGameStart(message.payload as GameStartPayload);
                        break;
                    case 'ROUND_START':
                        handleRoundStart(message.payload as RoundStartPayload);
                        break;
                    case 'ROUND_RESULT':
                        handleRoundResult(message.payload as RoundResultPayload);
                        break;
                    case 'GAME_OVER':
                        handleGameOver(message.payload as GameOverPayload);
                        break;
                    case 'ERROR':
                        handleError(message.payload as ErrorPayload);
                        break;
                    default:
                        console.warn('⚠️ Unknown message type:', message.type);
                }
            } catch (error) {
                console.error('❌ Failed to parse WebSocket message:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('❌ WebSocket error:', error);
            setIsConnected(false);
            hasConnectedRef.current = false;
        };

        ws.onclose = (event) => {
            console.log('🔌 WebSocket closed:', event.code, event.reason);
            setIsConnected(false);
            wsRef.current = null;
            hasConnectedRef.current = false;
        };

        // Cleanup on unmount
        return () => {
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                console.log('🧹 Cleaning up WebSocket connection');
                wsRef.current.close();
            }
        };
    }, [roomId, userId, handleGameStart, handleRoundStart, handleRoundResult, handleGameOver, handleError]);

    // ✅ Send answer function
    const sendAnswer = useCallback((color: StroopColor) => {
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            console.log('🎮 Sending answer:', color);
            
            const clickPayload: ClickPayload = {
                answer: color
            };

            const message: WSMessage = {
                type: 'CLICK',
                payload: clickPayload
            };

            wsRef.current.send(JSON.stringify(message));
        } else {
            console.error('❌ Cannot send answer: WebSocket not connected');
        }
    }, []);

    return {
        gameState,
        sendAnswer,
        isConnected,
    };
}