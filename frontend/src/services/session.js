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
}

export default new SessionService(); 