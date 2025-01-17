import { API_ENDPOINTS } from './api';
import { authService } from './auth';

class SessionService {
    async joinSession(sessionId) {
        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.JOIN, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ session_id: sessionId })
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

    async checkSessionMembership(sessionId) {
        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.CHECK_MEMBERSHIP(sessionId), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            return response.ok;
        } catch (error) {
            console.error('Error checking session membership:', error);
            return false;
        }
    }
}

export default new SessionService(); 