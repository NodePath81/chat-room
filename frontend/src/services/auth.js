import { API_ENDPOINTS } from './api';

class AuthService {
  setToken(token) {
    if (!token) {
      console.warn('Attempted to set null/undefined token');
      return;
    }
    console.log('Setting token:', token);
    localStorage.setItem('token', token);
  }

  getToken() {
    return localStorage.getItem('token');
  }

  setUser(user) {
    if (!user) {
      console.warn('Attempted to set null/undefined user');
      return;
    }
    console.log('Setting user:', user);
    localStorage.setItem('user', JSON.stringify(user));
  }

  getUser() {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr);
    } catch (error) {
      console.error('Error parsing user data:', error);
      return null;
    }
  }

  clearToken() {
    localStorage.removeItem('token');
  }

  clearUser() {
    localStorage.removeItem('user');
  }

  isAuthenticated() {
    return !!this.getToken();
  }

  logout() {
    this.clearToken();
    this.clearUser();
  }

  async fetchUserData() {
    const user = this.getUser();
    if (!user) return null;

    try {
      const response = await fetch(API_ENDPOINTS.USERS.GET(user.id), {
        headers: {
          'Authorization': `Bearer ${this.getToken()}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user data');
      }

      const userData = await response.json();
      this.setUser(userData);
      return userData;
    } catch (error) {
      console.error('Error fetching user data:', error);
      return null;
    }
  }

  updateStoredUser(userData) {
    if (!userData) {
      console.warn('Attempted to update with null/undefined user data');
      return;
    }
    this.setUser(userData);
  }
}

export const authService = new AuthService(); 