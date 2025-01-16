import { API_ENDPOINTS } from '../config';

export const authService = {
  async register(username, password) {
    const response = await fetch(API_ENDPOINTS.AUTH.REGISTER, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) {
      throw new Error('Registration failed');
    }
  },

  async login(username, password) {
    const response = await fetch(API_ENDPOINTS.AUTH.LOGIN, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) {
      throw new Error('Login failed');
    }
    const data = await response.json();
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));

    // Fetch latest user data including avatar URL
    await this.fetchUserData();
    return data;
  },

  logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  getToken() {
    return localStorage.getItem('token');
  },

  getUser() {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },

  isAuthenticated() {
    return !!this.getToken();
  },

  updateStoredUser(userData) {
    localStorage.setItem('user', JSON.stringify(userData));
  },

  async fetchUserData() {
    const user = this.getUser();
    if (!user) return null;

    try {
      const response = await fetch(`${API_ENDPOINTS.USERS.GET(user.id)}`, {
        headers: {
          'Authorization': `Bearer ${this.getToken()}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user data');
      }

      const userData = await response.json();
      this.updateStoredUser(userData);
      return userData;
    } catch (error) {
      console.error('Error fetching user data:', error);
      return null;
    }
  }
}; 