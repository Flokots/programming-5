const USER_SERVICE_URL = 'http://localhost:8001';
const ROOM_SERVICE_URL = 'http://localhost:8002';


// User Service APIs
/**
 * Register a new user
 * POST http://localhost:8001/register
 * Body: { "username": "alice" }
 * Returns: { "user_id": "uuid" }
 */
export async function registerUser(username: string): Promise<string> {
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
    return data.user_id;
  } catch (error) {
    console.error('Register error:', error);
    throw error;
  }
}

/**
 * Login existing user
 * POST http://localhost:8001/login
 * Body: { "username": "alice" }
 * Returns: { "user_id": "uuid" }
 */
export async function loginUser(username: string): Promise<string> {
  try {
    const response = await fetch(`${USER_SERVICE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      // User not found - will register instead
      if (response.status === 404) {
        throw new Error('USER_NOT_FOUND');
      }
      throw new Error(`Login failed: ${response.statusText}`);
    }

    const data = await response.json();
    return data.user_id;
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}

/**
 * Login or register user (convenience function)
 * Tries login first, registers if user doesn't exist
 */
export async function loginOrRegister(username: string): Promise<string> {
  try {
    // Try login first
    return await loginUser(username);
  } catch (error) {
    // If user not found, register
    if (error instanceof Error && error.message === 'USER_NOT_FOUND') {
      console.log('User not found, registering...');
      return await registerUser(username);
    }
    throw error;
  }
}


// Room Service APIs
/**
 * Join matchmaking queue
 * POST http://localhost:8002/join
 * Body: { "user_id": "uuid" }
 * Returns: { "room_id": "uuid", "player1_id": "uuid", "player2_id": "uuid" | null }
 */
export async function joinRoom(userId: string): Promise<string> {
  try {
    const response = await fetch(`${ROOM_SERVICE_URL}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to join room: ${response.statusText}`);
    }

    const data = await response.json();
    return data.room_id;
  } catch (error) {
    console.error('Join room error:', error);
    throw error;
  }
}

/**
 * Check if game is ready (both players joined)
 * GET http://localhost:8002/room/{roomId}/ready
 * Returns: { "ready": true/false, "player1_id": "uuid", "player2_id": "uuid" }
 */
export async function checkGameReady(roomId: string): Promise<boolean> {
  try {
    const response = await fetch(`${ROOM_SERVICE_URL}/room/${roomId}/ready`);

    if (!response.ok) {
      throw new Error(`Failed to check game status: ${response.statusText}`);
    }

    const data = await response.json();
    return data.ready;
  } catch (error) {
    console.error('Check game ready error:', error);
    throw error;
  }
}

/**
 * Wait for game to be ready (polls every second)
 * Returns when both players have joined
 */
export async function waitForGameReady(roomId: string, maxAttempts: number = 30): Promise<void> {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      const ready = await checkGameReady(roomId);
      if (ready) {
        console.log('Game is ready!');
        return;
      }
      console.log(`Waiting for opponent... (${i + 1}/${maxAttempts})`);
      await sleep(1000); // Wait 1 second
    } catch (error) {
      console.error('Error checking game status:', error);
    }
  }
  throw new Error('Timeout waiting for opponent');
}


// Helper Functions
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}