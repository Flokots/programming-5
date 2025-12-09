interface BackendAuthResponse {
  id?: string;
  user_id?: string;
  token: string;
  username: string;
}

export interface AuthResponse {
  user_id: string;
  token: string;
  username: string;
}

export class APIClient {
  private userServiceURL = 'http://192.168.30.152:8001';
  private roomServiceURL = 'http://192.168.30.152:8002';
  private gameServiceURL = 'http://192.168.30.152:8003';
  
  private token: string | null = null;
  private username: string | null = null;
  private userID: string | null = null;

  async login(username: string, password: string): Promise<AuthResponse> {
    const response = await fetch(`${this.userServiceURL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Login failed: ${error}`);
    }

    const rawData: BackendAuthResponse = await response.json();
    const data: AuthResponse = {
      user_id: rawData.user_id || rawData.id || '',
      token: rawData.token,
      username: rawData.username,
    };

    if (!data.user_id) {
      throw new Error('Backend did not return user ID');
    }

    this.token = data.token;
    this.username = data.username;
    this.userID = data.user_id;
    
    console.log(`✅ Logged in as: ${data.username} (ID: ${data.user_id})`);
    return data;
  }

  async register(username: string, password: string): Promise<AuthResponse> {
    const response = await fetch(`${this.userServiceURL}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Registration failed: ${error}`);
    }

    const rawData: BackendAuthResponse = await response.json();
    const data: AuthResponse = {
      user_id: rawData.user_id || rawData.id || '',
      token: rawData.token,
      username: rawData.username,
    };

    if (!data.user_id) {
      throw new Error('Backend did not return user ID');
    }

    this.token = data.token;
    this.username = data.username;
    this.userID = data.user_id;
    
    console.log(`✅ Registered new user: ${data.username} (ID: ${data.user_id})`);
    return data;
  }

  async joinMatchmaking(userId: string): Promise<{ room_id: string; status: string }> {
    if (!this.token) {
      throw new Error('Not authenticated - no token available');
    }

    console.log(`🎮 Joining matchmaking...`);

    const response = await fetch(`${this.roomServiceURL}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
      },
      body: JSON.stringify({ user_id: userId }),
    });

    if (!response.ok) {
      const error = await response.text();
      console.error(`❌ Join matchmaking failed (${response.status}):`, error);
      throw new Error(`Failed to join matchmaking: ${error}`);
    }

    const data = await response.json();
    console.log(`✅ Joined matchmaking. Room ID: ${data.room_id}`);
    return data;
  }

  async checkRoomReady(roomId: string): Promise<boolean> {
    const response = await fetch(`${this.roomServiceURL}/room/${roomId}/ready`);
    
    if (!response.ok) {
      console.warn(`⚠️  Failed to check room status: ${response.status}`);
      return false;
    }

    const data = await response.json();
    return data.ready;
  }

  async leaveRoom(roomId: string): Promise<void> {
    if (!this.token) return;

    try {
      await fetch(`${this.roomServiceURL}/rooms/${roomId}/leave`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.token}`,
        },
      });
      console.log(`✅ Left room: ${roomId}`);
    } catch (error) {
      console.error('❌ Error leaving room:', error);
    }
  }

  isAuthenticated(): boolean {
    return this.token !== null && this.userID !== null;
  }
}