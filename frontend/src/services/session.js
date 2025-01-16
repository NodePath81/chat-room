import { API_ENDPOINTS } from '../config';

class SessionService {
    async joinSession(sessionId) {
        const token = localStorage.getItem('token');
        if (!token) return false;

        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.JOIN(sessionId), {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
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

    async checkSessionMembership(sessionId) {
        const token = localStorage.getItem('token');
        if (!token) return false;

        try {
            const response = await fetch(API_ENDPOINTS.SESSIONS.CHECK(sessionId), {
                headers: {
                    'Authorization': `Bearer ${token}`
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