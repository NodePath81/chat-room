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
        credentials: options.credentials || 'omit', // Default to not sending cookies
    };

    const response = await fetch(url, { ...defaultOptions, ...options });
    return handleResponse(response);
};

// Utility function to make session-specific API requests
const makeSessionRequest = async (url, sessionId, options = {}) => {
    const sessionToken = await getSessionToken(sessionId);
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
            'Session-Token': sessionToken,
            ...options.headers,
        },
        credentials: 'omit', // Don't send cookies
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
        BATCH: `${API_BASE_URL}/api/users/batch`,
        UPDATE_NICKNAME: (id) => {
            if (!isValidUUID(id)) throw new Error('Invalid user ID');
            return `${API_BASE_URL}/api/users/${id}/nickname`;
        },
        UPDATE_USERNAME: (id) => {
            if (!isValidUUID(id)) throw new Error('Invalid user ID');
            return `${API_BASE_URL}/api/users/${id}/username`;
        },
    },
    SESSIONS: {
        GET_IDS: `${API_BASE_URL}/api/sessions/ids`,
        CREATE: `${API_BASE_URL}/api/sessions`,
        JOIN: (token) => `${API_BASE_URL}/api/sessions/join?token=${token}`,
        GET: `${API_BASE_URL}/api/sessions/session`,
        CHECK_ROLE: `${API_BASE_URL}/api/sessions/role`,
        GET_USERS_IDS: `${API_BASE_URL}/api/sessions/users/ids`,
        GET_MESSAGES_IDS: (params) => {
            const url = new URL(`${API_BASE_URL}/api/sessions/messages/ids`);
            if (params?.before) url.searchParams.set('before', params.before);
            if (params?.limit) url.searchParams.set('limit', params.limit);
            return url.toString();
        },
        FETCH_MESSAGES: `${API_BASE_URL}/api/sessions/messages/batch`,
        KICK_MEMBER: (memberId) => `${API_BASE_URL}/api/sessions/kick?memberId=${memberId}`,
        REMOVE: `${API_BASE_URL}/api/sessions`,
        CREATE_SHARE_LINK: `${API_BASE_URL}/api/sessions/share`,
        GET_SHARE_INFO: (token) => `${API_BASE_URL}/api/sessions/share/info?token=${token}`,
        UPLOAD_MESSAGE_IMAGE: `${API_BASE_URL}/api/sessions/messages/upload`,
        GET_TOKEN: `${API_BASE_URL}/api/sessions/token`,
        REFRESH_TOKEN: `${API_BASE_URL}/api/sessions/token/refresh`,
        REVOKE_TOKEN: `${API_BASE_URL}/api/sessions/token`,
        GET_WS_TOKEN: `${API_BASE_URL}/api/sessions/wstoken`,
        LEAVE: `${API_BASE_URL}/api/sessions/leave`,
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (wsToken) => `ws://localhost:8080/ws?token=${wsToken}`,
    },
};

// Session token management
const getSessionToken = async (sessionId) => {
    // Try to get token from localStorage
    const storageKey = `session_token_${sessionId}`;
    const storedData = localStorage.getItem(storageKey);
    
    if (storedData) {
        try {
            // Parse the stored data
            const tokenData = JSON.parse(storedData);
            
            // Check if token exists and not expired
            if (tokenData.token && tokenData.expires_at) {
                const expiresAt = new Date(tokenData.expires_at);
                // Add 5 seconds buffer to prevent edge cases
                if (expiresAt > new Date(Date.now() + 5000)) {
                    return tokenData.token;
                }
            }
        } catch (error) {
            console.debug('Invalid stored data format, requesting new token');
        }
        // Remove expired or invalid token
        localStorage.removeItem(storageKey);
    }

    // Request new token
    console.debug('Requesting new session token');
    const response = await makeRequest(API_ENDPOINTS.SESSIONS.GET_TOKEN + `?session_id=${sessionId}`);
    
    // Store the complete token data
    localStorage.setItem(storageKey, JSON.stringify({
        session_id: response.session_id,
        token: response.token,
        expires_at: response.expires_at
    }));

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
        batchGet: (ids) => makeRequest(API_ENDPOINTS.USERS.BATCH, {
            method: 'POST',
            body: JSON.stringify({ ids }),
        }),
        updateNickname: (id, nickname) => makeRequest(API_ENDPOINTS.USERS.UPDATE_NICKNAME(id), {
            method: 'PUT',
            body: JSON.stringify({ nickname }),
        }),
        updateUsername: (id, username) => makeRequest(API_ENDPOINTS.USERS.UPDATE_USERNAME(id), {
            method: 'PUT',
            body: JSON.stringify({ username }),
        }),
    },
    sessions: {
        // Public routes (require only auth token)
        getSessionIDs: () => makeRequest(API_ENDPOINTS.SESSIONS.GET_IDS),
        create: (data) => makeRequest(API_ENDPOINTS.SESSIONS.CREATE, {
            method: 'POST',
            body: JSON.stringify(data),
        }),
        join: (token) => makeRequest(API_ENDPOINTS.SESSIONS.JOIN(token), {
            method: 'POST'
        }),
        getShareInfo: (token) => makeRequest(API_ENDPOINTS.SESSIONS.GET_SHARE_INFO(token)),
        getToken: (sessionId) => makeRequest(API_ENDPOINTS.SESSIONS.GET_TOKEN + `?session_id=${sessionId}`),

        // Protected routes (require session token)
        get: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.GET, sessionId),
        checkRole: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.CHECK_ROLE, sessionId),
        getUserIDs: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.GET_USERS_IDS, sessionId),
        getMessageIDs: (sessionId, params) => makeSessionRequest(API_ENDPOINTS.SESSIONS.GET_MESSAGES_IDS(params), sessionId),
        fetchMessages: (sessionId, messageIDs) => makeSessionRequest(API_ENDPOINTS.SESSIONS.FETCH_MESSAGES, sessionId, {
            method: 'POST',
            body: JSON.stringify({ ids: messageIDs }),
        }),
        kickMember: (sessionId, memberId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.KICK_MEMBER(memberId), sessionId, {
            method: 'POST'
        }),
        remove: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.REMOVE, sessionId, {
            method: 'DELETE'
        }),
        uploadMessageImage: (sessionId, formData) => makeSessionRequest(API_ENDPOINTS.SESSIONS.UPLOAD_MESSAGE_IMAGE, sessionId, {
            method: 'POST',
            body: formData,
        }),
        refreshToken: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.REFRESH_TOKEN, sessionId),
        revokeToken: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.REVOKE_TOKEN, sessionId, {
            method: 'DELETE'
        }),
        getWsToken: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.GET_WS_TOKEN, sessionId),
        leave: (sessionId) => makeSessionRequest(API_ENDPOINTS.SESSIONS.LEAVE, sessionId, {
            method: 'POST'
        }),
        createShareLink: (sessionId, data) => makeSessionRequest(API_ENDPOINTS.SESSIONS.CREATE_SHARE_LINK, sessionId, {
            method: 'POST',
            body: JSON.stringify(data),
        }),
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