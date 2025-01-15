class SessionService {
    async joinSession(sessionId) {
        const token = localStorage.getItem('token');
        if (!token) return false;

        try {
            const response = await fetch(`http://localhost:8080/api/sessions/${sessionId}/join`, {
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
            const response = await fetch(`http://localhost:8080/api/sessions/${sessionId}/check`, {
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