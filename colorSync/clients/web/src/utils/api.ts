const USER_SERVICE_URL = 'http://localhost:8001';
const ROOM_SERVICE_URL = 'http://localhost:8002';

// User Service APIs
export async function registerUser(username: string): Promise<{ user_id: string; username: string }> {
  try {
    const response = await fetch(`${USER_SERVICE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      throw new Error(`Registration failed: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      user_id: data.id,
      username: data.username,
    };
  } catch (error) {
    console.error('Registration error:', error);
    throw error;
  }
}

export async function loginUser(username: string): Promise<{user_id: string; username: string }> {
  try {
    const response = await fetch(`${USER_SERVICE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('User not found');
      }
      throw new Error(`Login failed: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      user_id: data.id,
      username: data.username,
    };
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}

// Room Service APIs
export async function joinMatchmaking(userId: string): Promise<{ room_id: string; players: string[]; status: string; message: string }> {
  try {
    const payload = { user_id: userId };
    console.log("Send join request:", payload);

    const response = await fetch(`${ROOM_SERVICE_URL}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('Join room failed:', response.status, errorText);
      throw new Error(`Failed to join room: ${response.statusText}`);
    }

    const data = await response.json();
    console.log('Join room response:', data);
    return data;
  } catch (error) {
    console.error('Join room error:', error);
    throw error;
  }
}

export async function isRoomReady(roomId: string): Promise<boolean> {
  try {
    const response = await fetch(`${ROOM_SERVICE_URL}/room/${roomId}/ready`);
    
    if (!response.ok) {
      throw new Error(`Failed to check room readiness: ${response.statusText}`);
    }
    const data = await response.json();
    return data.ready;
  } catch (error) {
    console.error('Check room readiness error:', error);
    throw error;
  }
}