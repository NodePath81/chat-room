// API configuration
const API_BASE_URL = 'http://localhost:8080'; // or whatever port your backend is running on

export const API_ENDPOINTS = {
    AUTH: {
        REGISTER: `${API_BASE_URL}/api/auth/register`,
        LOGIN: `${API_BASE_URL}/api/auth/login`,
        CHECK_USERNAME: `${API_BASE_URL}/api/auth/check-username`,
        CHECK_NICKNAME: `${API_BASE_URL}/api/auth/check-nickname`,
    },
    USERS: {
        GET: (id) => `${API_BASE_URL}/api/users/${id}`,
        UPDATE_NICKNAME: (id) => `${API_BASE_URL}/api/users/${id}/nickname`,
        UPDATE_USERNAME: (id) => `${API_BASE_URL}/api/users/${id}/username`,
    },
    SESSIONS: {
        LIST: `${API_BASE_URL}/api/sessions`,
        CREATE: `${API_BASE_URL}/api/sessions`,
        JOIN: (id) => `${API_BASE_URL}/api/sessions/${id}/join`,
        LEAVE: (id) => `${API_BASE_URL}/api/sessions/${id}/leave`,
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (sessionId) => `ws://localhost:8080/ws?session_id=${sessionId}`,
    },
}; 