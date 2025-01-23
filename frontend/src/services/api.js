import { authService } from "./auth";

// API configuration
const API_BASE_URL = 'http://localhost:8080'; // or whatever port your backend is running on

// Utility function to validate UUID
const isValidUUID = (uuid) => {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
};

// Error class for API errors
class APIError extends Error {
    constructor(message, status, data) {
        super(message);
        this.name = 'APIError';
        this.status = status;
        this.data = data;
    }
}

// Utility function to handle API responses
const handleResponse = async (response) => {
    const contentType = response.headers.get('content-type');
    const isJson = contentType && contentType.includes('application/json');
    const data = isJson ? await response.json() : await response.text();

    if (!response.ok) {
        throw new APIError(
            data.message || 'An error occurred',
            response.status,
            data
        );
    }

    return data;
};

// Utility function to make API requests
const makeRequest = async (url, options = {}) => {
    const token = localStorage.getItem('token');
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            ...(token && { Authorization: `Bearer ${token}` }),
            ...options.headers,
        },
        credentials: 'include', // This enables sending and receiving cookies
    };

    const response = await fetch(url, { ...defaultOptions, ...options });
    return handleResponse(response);
};

// Utility function to validate session ID
const validateSessionId = (id) => {
    if (!id || !isValidUUID(id)) {
        throw new Error('Invalid session ID');
    }
    return id;
};

// API endpoints configuration
const API_ENDPOINTS = {
    AUTH: {
        LOGIN: `${API_BASE_URL}/api/auth/login`,
        REGISTER: `${API_BASE_URL}/api/auth/register`,
        CHECK_USERNAME: (username) => `${API_BASE_URL}/api/auth/check-username?username=${username}`,
        CHECK_NICKNAME: (nickname) => `${API_BASE_URL}/api/auth/check-nickname?nickname=${nickname}`,
    },
    USERS: {
        GET: (id) => {
            if (!isValidUUID(id)) throw new Error('Invalid user ID');
            return `${API_BASE_URL}/api/users/${id}`;
        },
        UPDATE_NICKNAME: (id) => {
            if (!isValidUUID(id)) throw new Error('Invalid user ID');
            return `${API_BASE_URL}/api/users/${id}/nickname`;
        },
        UPDATE_USERNAME: (id) => {
            if (!isValidUUID(id)) throw new Error('Invalid user ID');
            return `${API_BASE_URL}/api/users/${id}/username`;
        },
        GET_SESSIONS: `${API_BASE_URL}/api/users/sessions`,
    },
    SESSIONS: {
        LIST: `${API_BASE_URL}/api/sessions`,
        CREATE: `${API_BASE_URL}/api/sessions`,
        JOIN: (token) => `${API_BASE_URL}/api/sessions/join?token=${token}`,
        GET: `${API_BASE_URL}/api/sessions/session`,
        CHECK_ROLE: `${API_BASE_URL}/api/sessions/role`,
        LIST_MEMBERS: `${API_BASE_URL}/api/sessions/members`,
        KICK_MEMBER: (memberId) => `${API_BASE_URL}/api/sessions/kick?memberId=${memberId}`,
        REMOVE: `${API_BASE_URL}/api/sessions/remove`,
        CREATE_SHARE_LINK: `${API_BASE_URL}/api/sessions/share`,
        GET_SHARE_INFO: (token) => `${API_BASE_URL}/api/sessions/share/info?token=${token}`,
        GET_MESSAGES: (params) => {
            const url = new URL(`${API_BASE_URL}/api/sessions/messages`);
            if (params?.before) url.searchParams.set('before', params.before);
            if (params?.limit) url.searchParams.set('limit', params.limit);
            return url.toString();
        },
        UPLOAD_MESSAGE_IMAGE: `${API_BASE_URL}/api/sessions/messages/upload`,
        GET_TOKEN: `${API_BASE_URL}/api/sessions/token`,
        REFRESH_TOKEN: `${API_BASE_URL}/api/sessions/token/refresh`,
        REVOKE_TOKEN: `${API_BASE_URL}/api/sessions/token`,
        GET_WS_TOKEN: `${API_BASE_URL}/api/sessions/wstoken`,
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (sessionId, wsToken) => {
            validateSessionId(sessionId);
            return `ws://localhost:8080/ws?token=${wsToken}`;
        },
    },
};

// Session token management
const getSessionToken = async (sessionId) => {
    // First try to get token from cookie
    // The cookie would be named session_token_{sessionId}
    const cookies = document.cookie.split(';');
    const tokenCookie = cookies.find(cookie => 
        cookie.trim().startsWith(`session_token_${sessionId}=`)
    );
    
    if (tokenCookie) {
        // If we have a cookie with the token, return it
        const token = tokenCookie.split('=')[1].trim();
        if (token) {
            return token;
        }
    }

    // If no valid token in cookie, request a new one
    const response = await makeRequest(API_ENDPOINTS.SESSIONS.GET_TOKEN + `?session_id=${sessionId}`);
    return response.token;
};

// API functions
export const api = {
    auth: {
        login: (credentials) => makeRequest(API_ENDPOINTS.AUTH.LOGIN, {
            method: 'POST',
            body: JSON.stringify(credentials),
        }),
        register: (userData) => makeRequest(API_ENDPOINTS.AUTH.REGISTER, {
            method: 'POST',
            body: JSON.stringify(userData),
        }),
        checkUsername: (username) => makeRequest(API_ENDPOINTS.AUTH.CHECK_USERNAME(username)),
        checkNickname: (nickname) => makeRequest(API_ENDPOINTS.AUTH.CHECK_NICKNAME(nickname)),
    },
    users: {
        get: (id) => makeRequest(API_ENDPOINTS.USERS.GET(id)),
        updateNickname: (id, nickname) => makeRequest(API_ENDPOINTS.USERS.UPDATE_NICKNAME(id), {
            method: 'PUT',
            body: JSON.stringify({ nickname }),
        }),
        updateUsername: (id, username) => makeRequest(API_ENDPOINTS.USERS.UPDATE_USERNAME(id), {
            method: 'PUT',
            body: JSON.stringify({ username }),
        }),
        getSessions: () => makeRequest(API_ENDPOINTS.USERS.GET_SESSIONS),
    },
    sessions: {
        // Public routes (require only auth token)
        list: () => makeRequest(API_ENDPOINTS.SESSIONS.LIST),
        create: (data) => makeRequest(API_ENDPOINTS.SESSIONS.CREATE, {
            method: 'POST',
            body: JSON.stringify(data),
        }),
        join: (token) => makeRequest(API_ENDPOINTS.SESSIONS.JOIN(token)),
        getShareInfo: (token) => makeRequest(API_ENDPOINTS.SESSIONS.GET_SHARE_INFO(token)),
        getToken: (sessionId) => makeRequest(API_ENDPOINTS.SESSIONS.GET_TOKEN + `?session_id=${sessionId}`),
        createShareLink: (data) => makeRequest(API_ENDPOINTS.SESSIONS.CREATE_SHARE_LINK, {
            method: 'POST',
            body: JSON.stringify(data),
        }),

        // Protected routes (require session token)
        // These methods will automatically get a session token if needed
        async get(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.GET);
        },
        async checkRole(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.CHECK_ROLE);
        },
        async listMembers(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.LIST_MEMBERS);
        },
        async kickMember(sessionId, memberId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.KICK_MEMBER(memberId));
        },
        async remove(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.REMOVE);
        },
        async getMessages(sessionId, params) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.GET_MESSAGES(params));
        },
        async uploadMessageImage(sessionId, formData) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.UPLOAD_MESSAGE_IMAGE, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                },
                body: formData,
            });
        },
        async refreshToken(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.REFRESH_TOKEN);
        },
        async revokeToken(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.REVOKE_TOKEN, {
                method: 'DELETE'
            });
        },
        async getWsToken(sessionId) {
            await getSessionToken(sessionId);
            return makeRequest(API_ENDPOINTS.SESSIONS.GET_WS_TOKEN);
        },
    },
    avatar: {
        upload: (formData) => makeRequest(API_ENDPOINTS.AVATAR.UPLOAD, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: formData,
        }),
    },
};

export { API_ENDPOINTS, isValidUUID, APIError }; 