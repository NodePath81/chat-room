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
        GET_SESSIONS: `${API_BASE_URL}/api/users/sessions`,
    },
    SESSIONS: {
        LIST: `${API_BASE_URL}/api/sessions`,
        CREATE: `${API_BASE_URL}/api/sessions`,
        GET: (id) => `${API_BASE_URL}/api/sessions/${id}`,
        JOIN: `${API_BASE_URL}/api/sessions/join`,
        CHECK_MEMBERSHIP: (id) => `${API_BASE_URL}/api/sessions/${id}`,
        CHECK_ROLE: (id) => `${API_BASE_URL}/api/sessions/${id}/role`,
        GET_MEMBERS: (id) => `${API_BASE_URL}/api/sessions/${id}/members`,
        KICK_MEMBER: (id) => `${API_BASE_URL}/api/sessions/${id}/kick`,
        REMOVE: (id) => `${API_BASE_URL}/api/sessions/${id}/remove`,
        CREATE_SHARE_LINK: (id) => `${API_BASE_URL}/api/sessions/${id}/share`,
        GET_SHARE_INFO: `${API_BASE_URL}/api/sessions/share/info`,
        GET_MESSAGES: (id, params) => {
            const url = new URL(`${API_BASE_URL}/api/sessions/${id}/messages`);
            if (params?.before) url.searchParams.set('before', params.before);
            if (params?.limit) url.searchParams.set('limit', params.limit);
            return url.toString();
        }
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (sessionId) => `ws://localhost:8080/ws?session_id=${sessionId}`,
    },
}; 