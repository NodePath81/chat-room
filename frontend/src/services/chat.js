import { API_ENDPOINTS } from './api';
import { authService } from './auth';
import { connect, disconnect, sendMessage, onMessage } from './websocket';

class ChatService {
    constructor() {
        this.messageHandlers = new Map();
        this.currentSessionId = null;
    }

    async connectToSession(sessionId) {
        this.currentSessionId = sessionId;
        connect(sessionId);
    }

    disconnectFromSession(sessionId) {
        disconnect(sessionId);
        this.currentSessionId = null;
    }

    async sendTextMessage(content) {
        if (!content.trim() || !this.currentSessionId) {
            throw new Error('Invalid message or no active session');
        }

        await sendMessage({
            content: content.trim(),
            type: 'text',
            session_id: this.currentSessionId
        });
    }

    async uploadImage(file) {
        if (!file || !this.currentSessionId) {
            throw new Error('No file selected or no active session');
        }

        const formData = new FormData();
        formData.append('image', file);
        formData.append('session_id', this.currentSessionId);
        formData.append('type', 'image');

        const response = await fetch(API_ENDPOINTS.SESSIONS.UPLOAD_MESSAGE_IMAGE(this.currentSessionId), {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error('Failed to upload image');
        }
    }

    async getMessages(beforeTimestamp = null, limit = 20) {
        if (!this.currentSessionId) {
            throw new Error('No active session');
        }

        const url = new URL(API_ENDPOINTS.SESSIONS.GET_MESSAGES(this.currentSessionId));
        if (beforeTimestamp) {
            url.searchParams.append('before', beforeTimestamp);
        }
        url.searchParams.append('limit', limit.toString());

        const response = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to get messages');
        }

        const data = await response.json();
        const messages = data.messages || [];
        messages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

        return {
            messages,
            hasMore: data.has_more || false,
            oldestTimestamp: messages[0]?.timestamp
        };
    }

    onMessageReceived(handler) {
        if (!this.currentSessionId) {
            throw new Error('No active session');
        }

        onMessage(handler, this.currentSessionId);
    }

    async fetchUserData(userId) {
        const response = await fetch(API_ENDPOINTS.USERS.GET(userId), {
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch user data');
        }

        return await response.json();
    }
}

export const chatService = new ChatService(); 