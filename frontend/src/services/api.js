// API configuration
const API_BASE_URL = 'http://localhost:8080'; // or whatever port your backend is running on

// Utility function to validate UUID
export const isValidUUID = (uuid) => {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
};

// Error class for API errors
export class APIError extends Error {
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

export const API_ENDPOINTS = {
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
        GET: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}`;
        },
        CHECK_ROLE: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}/role`;
        },
        LIST_MEMBERS: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}/members`;
        },
        KICK_MEMBER: (sessionId, memberId) => {
            validateSessionId(sessionId);
            if (!isValidUUID(memberId)) throw new Error('Invalid member ID');
            return `${API_BASE_URL}/api/sessions/${sessionId}/kick?memberId=${memberId}`;
        },
        REMOVE: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}/remove`;
        },
        CREATE_SHARE_LINK: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}/share`;
        },
        GET_SHARE_INFO: `${API_BASE_URL}/api/sessions/share/info`,
        GET_MESSAGES: (id, params) => {
            validateSessionId(id);
            const url = new URL(`${API_BASE_URL}/api/sessions/${id}/messages`);
            if (params?.before) url.searchParams.set('before', params.before);
            if (params?.limit) url.searchParams.set('limit', params.limit);
            return url.toString();
        },
        UPLOAD_MESSAGE_IMAGE: (id) => {
            validateSessionId(id);
            return `${API_BASE_URL}/api/sessions/${id}/messages/upload`;
        },
    },
    AVATAR: {
        UPLOAD: `${API_BASE_URL}/api/avatar`,
    },
    WEBSOCKET: {
        CONNECT: (sessionId) => {
            validateSessionId(sessionId);
            return `ws://localhost:8080/ws?sessionId=${sessionId}`;
        },
    },
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
        list: () => makeRequest(API_ENDPOINTS.SESSIONS.LIST),
        create: (data) => makeRequest(API_ENDPOINTS.SESSIONS.CREATE, {
            method: 'POST',
            body: JSON.stringify(data),
        }),
        join: (token) => makeRequest(API_ENDPOINTS.SESSIONS.JOIN(token)),
        get: (id) => makeRequest(API_ENDPOINTS.SESSIONS.GET(id)),
        checkRole: (id) => makeRequest(API_ENDPOINTS.SESSIONS.CHECK_ROLE(id)),
        listMembers: (id) => makeRequest(API_ENDPOINTS.SESSIONS.LIST_MEMBERS(id)),
        kickMember: (sessionId, userId) => makeRequest(API_ENDPOINTS.SESSIONS.KICK_MEMBER(sessionId, userId)),
        remove: (id) => makeRequest(API_ENDPOINTS.SESSIONS.REMOVE(id)),
        createShareLink: (id, data) => makeRequest(API_ENDPOINTS.SESSIONS.CREATE_SHARE_LINK(id), {
            method: 'POST',
            body: JSON.stringify(data),
        }),
        getShareInfo: () => makeRequest(API_ENDPOINTS.SESSIONS.GET_SHARE_INFO),
        getMessages: (id, params) => makeRequest(API_ENDPOINTS.SESSIONS.GET_MESSAGES(id, params)),
        uploadMessageImage: (id, formData) => makeRequest(API_ENDPOINTS.SESSIONS.UPLOAD_MESSAGE_IMAGE(id), {
            method: 'POST',
            headers: {}, // Let browser set content-type for multipart/form-data
            body: formData,
        }),
    },
    avatar: {
        upload: (formData) => makeRequest(API_ENDPOINTS.AVATAR.UPLOAD, {
            method: 'POST',
            headers: {}, // Let browser set content-type for multipart/form-data
            body: formData,
        }),
    },
}; 