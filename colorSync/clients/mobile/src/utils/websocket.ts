import type { WSMessage, ClickMessage, ColorOption } from '../types';

export class GameWebSocket {
  private ws: WebSocket | null = null;
  private messageHandlers: Map<string, (payload: unknown) => void> = new Map();

  connect(roomId: string, userId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const url = `ws://localhost:8003/game/ws?room_id=${roomId}&user_id=${userId}`;
      
      console.log('🔌 Connecting to game via WebSocket...');
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        console.log('✅ Connected to game via WebSocket');
        resolve();
      };

      this.ws.onerror = (error) => {
        console.error('❌ WebSocket error:', error);
        reject(new Error('Failed to connect to game server'));
      };

      this.ws.onclose = (event) => {
        console.log(`🔌 WebSocket closed: ${event.code}`);
        this.ws = null;
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          console.log(`📨 Received: ${message.type}`, message.payload);
          
          const handler = this.messageHandlers.get(message.type);
          if (handler) {
            handler(message.payload);
          }
        } catch (error) {
          console.error('❌ Error parsing WebSocket message:', error);
        }
      };
    });
  }

  disconnect(): void {
    if (this.ws) {
      console.log('🔌 Disconnecting WebSocket...');
      this.ws.close();
      this.ws = null;
    }
    this.messageHandlers.clear();
  }

  on<T = unknown>(messageType: string, handler: (payload: T) => void): void {
    this.messageHandlers.set(messageType, handler as (payload: unknown) => void);
  }

  off(messageType: string): void {
    this.messageHandlers.delete(messageType);
  }

  sendClick(color: ColorOption): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('❌ Cannot send click: WebSocket not connected');
      return;
    }

    const message: ClickMessage = {
      type: 'CLICK',
      payload: { answer: color },
    };

    this.ws.send(JSON.stringify(message));
    console.log(`🖱️  Clicked: ${color}`);
  }
}