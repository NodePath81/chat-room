import { websocketService } from './websocket';
import { api } from './api';

class ChatService {
    constructor() {
        this.currentSessionId = null;
    }

    async connectToSession(sessionId) {
        this.currentSessionId = sessionId;
        websocketService.connect(sessionId);
    }

    disconnectFromSession() {
        if (this.currentSessionId) {
            websocketService.disconnect();
            this.currentSessionId = null;
        }
    }

    onMessageReceived(callback) {
        websocketService.onMessage(callback);
    }

    async sendTextMessage(content) {
        if (!this.currentSessionId) {
            throw new Error('Not connected to any session');
        }

        websocketService.sendMessage({
            type: 'text',
            content: content.trim(),
            session_id: this.currentSessionId
        });
    }

    async fetchUserData(userId) {
        return api.users.get(userId);
    }
}

export const chatService = new ChatService(); 