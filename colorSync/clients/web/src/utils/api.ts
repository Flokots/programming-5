const USER_SERVICE_URL = 'https://localhost:8001';
const ROOM_SERVICE_URL = 'https://localhost:8002';

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
    return data;
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
    return data;
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}