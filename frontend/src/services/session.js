import { api } from './api';

class SessionService {
    async createSession(name) {
        try {
            const response = await api.sessions.create({ name });
            return response;
        } catch (error) {
            console.error('Error creating session:', error);
            throw error;
        }
    }

    async joinSession(token) {
        try {
            await api.sessions.join(token);
            return true;
        } catch (error) {
            console.error('Error joining session:', error);
            return false;
        }
    }

    async checkSessionRole(sessionId) {
        try {
            const response = await api.sessions.checkRole(sessionId);
            return response.role;
        } catch (error) {
            console.error('Error checking session role:', error);
            return null;
        }
    }

    async getSessionIDs() {
        try {
            const response = await api.sessions.getSessionIDs();
            return response.session_ids || [];
        } catch (error) {
            console.error('Error getting session IDs:', error);
            throw error;
        }
    }

    async getSession(sessionId) {
        try {
            return await api.sessions.get(sessionId);
        } catch (error) {
            console.error('Error getting session details:', error);
            throw error;
        }
    }

    async getSessions(sessionIds) {
        try {
            if (!sessionIds || sessionIds.length === 0) {
                return [];
            }

            const sessionsData = [];
            for (const sessionId of sessionIds) {
                try {
                    const sessionData = await this.getSession(sessionId);
                    const roleData = await this.checkSessionRole(sessionId);
                    sessionsData.push({
                        ...sessionData,
                        userRole: roleData
                    });
                } catch (error) {
                    console.error(`Error fetching session ${sessionId}:`, error);
                }
            }
            return sessionsData;
        } catch (error) {
            console.error('Error getting sessions:', error);
            throw error;
        }
    }

    async leaveSession(sessionId) {
        try {
            await api.sessions.leave(sessionId);
            return true;
        } catch (error) {
            console.error('Error leaving session:', error);
            throw error;
        }
    }

    async getSessionUserIDs(sessionId) {
        try {
            const response = await api.sessions.getUserIDs(sessionId);
            return response.user_ids || [];
        } catch (error) {
            console.error('Error getting session user IDs:', error);
            throw error;
        }
    }

    async getSessionUsers(sessionId) {
        try {
            const userIds = await this.getSessionUserIDs(sessionId);
            if (!userIds || userIds.length === 0) {
                return [];
            }

            const response = await api.users.batchGet(userIds);
            return response.users || [];
        } catch (error) {
            console.error('Error getting session users:', error);
            throw error;
        }
    }

    async getWebSocketToken(sessionId) {
        try {
            const response = await api.sessions.getWsToken(sessionId);
            return response.token;
        } catch (error) {
            console.error('Error getting WebSocket token:', error);
            throw error;
        }
    }

    async initializeSession(sessionId) {
        try {
            // Get session token first
            await api.sessions.getToken(sessionId);

            // Check user role
            const roleResponse = await api.sessions.checkRole(sessionId);
            if (!roleResponse || !roleResponse.role) {
                throw new Error('Failed to get user role');
            }

            // Get session details
            const sessionDetails = await api.sessions.get(sessionId);
            if (!sessionDetails) {
                throw new Error('Failed to get session details');
            }

            return {
                role: roleResponse.role,
                session: sessionDetails,
                isCreator: roleResponse.role === 'creator'
            };
        } catch (error) {
            console.error('Error initializing session:', error);
            throw error;
        }
    }

    async getMessages(sessionId, before = null, limit = 50) {
        try {
            const response = await api.sessions.getMessageIDs(sessionId, {
                before,
                limit: limit + 1 // Get one extra to check if there are more
            });

            if (!response.message_ids || response.message_ids.length === 0) {
                return {
                    messages: [],
                    hasMore: false
                };
            }

            const hasMore = response.message_ids.length > limit;
            const messageIds = hasMore ? response.message_ids.slice(0, limit) : response.message_ids;

            const messagesResponse = await api.sessions.fetchMessages(sessionId, messageIds);
            const messages = messagesResponse.messages || [];

            // Sort messages by timestamp
            messages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

            return {
                messages,
                hasMore
            };
        } catch (error) {
            console.error('Error getting messages:', error);
            throw error;
        }
    }

    async createShareLink(sessionId, durationDays) {
        try {
            const response = await api.sessions.createShareLink(sessionId, { durationDays });
            return response;
        } catch (error) {
            console.error('Error creating share link:', error);
            throw error;
        }
    }

    async kickMember(sessionId, memberId) {
        try {
            await api.sessions.kickMember(sessionId, memberId);
        } catch (error) {
            console.error('Error kicking member:', error);
            throw error;
        }
    }

    async removeSession(sessionId) {
        try {
            await api.sessions.remove(sessionId);
        } catch (error) {
            console.error('Error removing session:', error);
            throw error;
        }
    }

    async getSessionMembers(sessionId) {
        try {
            const userIds = await this.getSessionUserIDs(sessionId);
            return userIds;
        } catch (error) {
            console.error('Error getting session members:', error);
            throw error;
        }
    }

    async uploadMessageImage(sessionId, imageFile) {
        try {
            const formData = new FormData();
            formData.append('image', imageFile);
            
            const response = await api.sessions.uploadMessageImage(sessionId, formData);
            return response;
        } catch (error) {
            console.error('Error uploading message image:', error);
            throw error;
        }
    }
}

export default new SessionService(); 