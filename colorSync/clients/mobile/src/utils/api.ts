// ⚠️ IMPORTANT: For React Native, localhost won't work on physical devices
// Use your computer's local IP address instead
// Find it with: ipconfig (Windows) or ifconfig (Mac/Linux)
const USER_SERVICE_URL = 'http://192.168.30.152:8001';  // Change to http://YOUR_IP:8001 for device testing
const ROOM_SERVICE_URL = 'http://192.168.30.152:8002';  // Change to http://YOUR_IP:8002 for device testing

/**
 * Register a new user
 */
export async function registerUser(username: string): Promise<{ user_id: string; username: string }> {
  try {
    console.log('📝 Registering user:', username);
    
    const response = await fetch(`${USER_SERVICE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Registration failed: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    console.log('✅ User registered:', data);
    
    return {
      user_id: data.id,
      username: data.username
    };
  } catch (error) {
    console.error('❌ Register error:', error);
    throw error;
  }
}

/**
 * Login existing user
 */
export async function loginUser(username: string): Promise<{ user_id: string; username: string }> {
  try {
    console.log('🔑 Logging in user:', username);
    
    const response = await fetch(`${USER_SERVICE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('User not found. Please register first.');
      }
      const errorText = await response.text();
      throw new Error(`Login failed: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    console.log('✅ User logged in:', data);
    
    return {
      user_id: data.id,
      username: data.username
    };
  } catch (error) {
    console.error('❌ Login error:', error);
    throw error;
  }
}

/**
 * Join matchmaking queue
 */
export async function joinMatchmaking(userId: string): Promise<{ 
  room_id: string; 
  players: string[]; 
  status: string; 
  message: string 
}> {
  try {
    console.log('🎮 Joining matchmaking for user:', userId);
    
    const response = await fetch(`${ROOM_SERVICE_URL}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userId }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to join matchmaking: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    console.log('✅ Joined room:', data);
    
    return data;
  } catch (error) {
    console.error('❌ Join matchmaking error:', error);
    throw error;
  }
}

/**
 * Check if room is ready (both players joined)
 */
export async function checkRoomReady(roomId: string): Promise<{ ready: boolean; players: string[] }> {
  try {
    const response = await fetch(`${ROOM_SERVICE_URL}/room/${roomId}/ready`);

    if (!response.ok) {
      throw new Error(`Failed to check room status: ${response.statusText}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('❌ Check room ready error:', error);
    throw error;
  }
}

/**
 * Get WebSocket URL for game connection
 */
export function getGameWebSocketUrl(roomId: string, userId: string): string {
  // ⚠️ Change to ws://YOUR_IP:8003 for device testing
  return `ws://localhost:8003/game/ws?room_id=${roomId}&user_id=${userId}`;
}