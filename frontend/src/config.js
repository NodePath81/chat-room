// Base URLs for API and WebSocket
export const API_BASE_URL = 'http://localhost:8080';
export const WS_BASE_URL = 'ws://localhost:8080';

// API endpoints
export const API_ENDPOINTS = {
    AUTH: {
        REGISTER: `${API_BASE_URL}/api/auth/register`,
        LOGIN: `${API_BASE_URL}/api/auth/login`,
    },
    SESSIONS: {
        LIST: `${API_BASE_URL}/api/sessions`,
        JOIN: (id) => `${API_BASE_URL}/api/sessions/${id}/join`,
        CHECK: (id) => `${API_BASE_URL}/api/sessions/${id}/check`,
    },
    USERS: {
        GET: (id) => `${API_BASE_URL}/api/users/${id}`,
    },
    WEBSOCKET: (sessionId) => `${WS_BASE_URL}/ws?sessionId=${sessionId}`,
}; 