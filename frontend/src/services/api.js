// API configuration
const API_BASE_URL = 'http://localhost:8080'; // or whatever port your backend is running on

export const API_ENDPOINTS = {
    AUTH: {
        LOGIN: `${API_BASE_URL}/api/auth/login`,
        REGISTER: `${API_BASE_URL}/api/auth/register`,
        CHECK_USERNAME: (username) => `${API_BASE_URL}/api/auth/check-username?username=${username}`,
        CHECK_NICKNAME: (nickname) => `${API_BASE_URL}/api/auth/check-nickname?nickname=${nickname}`,
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
        JOIN: (token) => `${API_BASE_URL}/api/sessions/join?token=${token}`,
        GET: (id) => `${API_BASE_URL}/api/sessions/${id}`,
        CHECK_ROLE: (id) => `${API_BASE_URL}/api/sessions/${id}/role`,
        LIST_MEMBERS: (id) => `${API_BASE_URL}/api/sessions/${id}/members`,
        KICK_MEMBER: (sessionId, userId) => `${API_BASE_URL}/api/sessions/${sessionId}/kick?userId=${userId}`,
        REMOVE: (id) => `${API_BASE_URL}/api/sessions/${id}/remove`,
        CREATE_SHARE_LINK: (id) => `${API_BASE_URL}/api/sessions/${id}/share`,
        GET_SHARE_INFO: `${API_BASE_URL}/api/sessions/share/info`,
        GET_MESSAGES: (id, params) => {
            const url = new URL(`${API_BASE_URL}/api/sessions/${id}/messages`);
            if (params?.before) url.searchParams.set('before', params.before);
            if (params?.limit) url.searchParams.set('limit', params.limit);
            return url.toString();
        },
        UPLOAD_MESSAGE_IMAGE: (id) => `${API_BASE_URL}/api/sessions/${id}/messages/upload`,
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (sessionId) => `ws://localhost:8080/ws?session_id=${sessionId}`,
    },
}; 