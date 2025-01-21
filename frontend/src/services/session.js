import { API_ENDPOINTS } from './api';
import { authService } from './auth';

class SessionService {
    async joinSession(token) {
        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.JOIN(token), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to join session');
            }

            return true;
        } catch (error) {
            console.error('Error joining session:', error);
            return false;
        }
    }

    async checkSessionRole(sessionId) {
        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.CHECK_ROLE(sessionId), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            if (!response.ok) {
                return null;
            }

            const data = await response.json();
            return data.role;
        } catch (error) {
            console.error('Error checking session role:', error);
            return null;
        }
    }

    async getSession(sessionId) {
        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.GET(sessionId), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to get session details');
            }

            return await response.json();
        } catch (error) {
            console.error('Error getting session details:', error);
            throw error;
        }
    }

    async getMessages(sessionId, beforeId = null, limit = 1) {
        try {
            const url = new URL(API_ENDPOINTS.SESSIONS.GET_MESSAGES(sessionId));
            if (beforeId) {
                url.searchParams.append('before', beforeId);
            }
            url.searchParams.append('limit', limit.toString());

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to get messages');
            }

            const data = await response.json();
            // Sort messages by timestamp in ascending order
            const messages = data.messages || [];
            messages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
            
            return {
                messages,
                hasMore: data.has_more || false
            };
        } catch (error) {
            console.error('Error getting messages:', error);
            throw error;
        }
    }
}

export default new SessionService(); 