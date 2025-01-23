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
            console.error('Not connected to any session');
            throw new Error('Not connected to any session');
        }

        console.debug('Sending text message:', {
            sessionId: this.currentSessionId,
            content: content,
            type: 'text'
        });

        websocketService.sendMessage({
            type: 'text',
            content: content.trim()
        });
    }

    async uploadImage(file) {
        if (!this.currentSessionId) {
            console.error('Not connected to any session');
            throw new Error('Not connected to any session');
        }

        console.debug('Uploading image:', {
            sessionId: this.currentSessionId,
            fileName: file.name,
            fileSize: file.size
        });

        try {
            
            // Create form data
            const formData = new FormData();
            formData.append('image', file);

            // Upload image
            const response = await api.sessions.uploadMessageImage(this.currentSessionId, formData);
            console.debug('Image upload successful:', response);

            return response;
        } catch (error) {
            console.error('Failed to upload image:', error);
            throw error;
        }
    }

    async fetchUserData(userId) {
        return api.users.get(userId);
    }
}

export const chatService = new ChatService(); 