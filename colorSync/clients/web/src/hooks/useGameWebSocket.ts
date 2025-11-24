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
    // Game state
    const [gameState, setGameState] = useState<GameState>({
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
        roomId,
    });

    const [isConnected, setIsConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);

    //  Define Message Handlers
    const handleGameStart = useCallback((payload: GameStartPayload) => {
        console.log('Game starting with', payload.max_rounds, 'rounds');
        setGameState(prev => ({
            ...prev,
            status: 'playing',
            maxRounds: payload.max_rounds,
            currentRound: 0,
            myScore: 0,
            opponentScore: 0,
        }));
    }, []);

    const handleRoundStart = useCallback((payload: RoundStartPayload) => {
        console.log(`Round ${payload.round}: ${payload.word} (${payload.color})`);
        setGameState(prev => ({
            ...prev,
            currentRound: payload.round,
            word: payload.word,
            color: payload.color,
        }));
    }, []);

    const handleRoundResult = useCallback((payload: RoundResultPayload) => {
        console.log('Round result - Winner:', payload.winner);
        console.log('Current user ID:', userId);

        setGameState(prev => {
            let newMyScore = prev.myScore;
            let newOpponentScore = prev.opponentScore;

            if (payload.winner === userId) {
                newMyScore += 1;
                console.log('I won this round! New myScore:', newMyScore);
            } else if (payload.winner !== null && payload.winner !== userId) {
                // Opponent won this round
                newOpponentScore += 1;
                console.log('Opponent won this round. New opponentScore:', newOpponentScore);
            } else {
                console.log('Round timed out - no winner');
            }

            return {
                ...prev,
                myScore: newMyScore,
                opponentScore: newOpponentScore,
            };
        });
    }, [userId]);

    const handleGameOver = useCallback((payload: GameOverPayload) => {
        console.log('Game over! Winner:', payload.winner);

        const myStats = payload.stats[userId] || null;

        setGameState(prev => ({
            ...prev,
            status: 'game_over',
            winner: payload.winner,
            myStats,
        }));
    }, [userId]);

    const handleWrongAnswer = useCallback(() => {
        console.log('Wrong answer! Blocked for this round');
        setGameState(prev => ({
            ...prev,
            errorMessage: 'Wrong answer! Wait for the next round.',
        }));

        setTimeout(() => {
            setGameState(prev => ({
                ...prev,
                errorMessage: null,
            }));
        }, 2000);
    }, []);

    const handleError = useCallback((payload: ErrorPayload) => {
        console.error('Server error:', payload.message);
        setGameState(prev => ({
            ...prev,
            status: 'error',
            errorMessage: payload.message,
        }));
    }, []);

    // WebSocket Effect
    useEffect(() => {
        // Don't connect if roomId or userId are empty
        if (!roomId || !userId) {
            console.log('Skipping WebSocket connection (no roomId or userId)');
            return;
        }

        console.log('Connecting to WebSocket:', roomId, userId);

        const ws = new WebSocket(
            `ws://localhost:8003/game/ws?room_id=${roomId}&user_id=${userId}`
        );

        wsRef.current = ws;

        // Connection opened
        ws.onopen = () => {
            console.log('WebSocket connected to game');
            setIsConnected(true);
            setGameState(prev => ({ ...prev, status: 'waiting' }));
        };

        ws.onmessage = (event) => {
            const msg: WSMessage = JSON.parse(event.data);
            console.log('Received message:', msg.type, msg.payload);

            switch (msg.type) {
                case "GAME_START":
                    handleGameStart(msg.payload as GameStartPayload);
                    break;

                case "ROUND_START":
                    handleRoundStart(msg.payload as RoundStartPayload);
                    break;

                case "ROUND_RESULT":
                    handleRoundResult(msg.payload as RoundResultPayload);
                    break;

                case "GAME_OVER":
                    handleGameOver(msg.payload as GameOverPayload);
                    break;

                case "WRONG_ANSWER":
                    handleWrongAnswer();
                    break;

                case "ERROR":
                    handleError(msg.payload as ErrorPayload);
                    break;

                default:
                    console.warn('Unknown message type:', msg.type);
            }
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected');
            setIsConnected(false);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            setGameState(prev => ({
                ...prev,
                status: 'error',
                errorMessage: 'Connection error occurred.',
            }));
        };

        return () => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.close();
            }
        };
    }, [roomId, userId, handleGameStart, handleRoundStart, handleRoundResult, handleGameOver, handleWrongAnswer, handleError]);

    // Send Answer Function
    const sendAnswer = (color: StroopColor) => {
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            console.log('Sending answer:', color);

            const clickPayload: ClickPayload = {
                answer: color
            };
            const message: WSMessage = {
                type: 'CLICK',
                payload: clickPayload
            };
            wsRef.current.send(JSON.stringify(message));
        } else {
            console.error('Cannot send answer: WebSocket not connected');
        }
    };


    // Return Hook Interface
    return {
        gameState,
        sendAnswer,
        isConnected,
    };
}