import { api } from './api';

class UserService {
    constructor() {
        this.userCache = new Map();
        this.pendingRequests = new Map();
    }

    // Get user data from cache or fetch from API
    async getUser(userId) {
        // Check cache first
        if (this.userCache.has(userId)) {
            return this.userCache.get(userId);
        }

        // Check if there's a pending request for this user
        if (this.pendingRequests.has(userId)) {
            return this.pendingRequests.get(userId);
        }

        // Create new request
        const promise = this._fetchUser(userId);
        this.pendingRequests.set(userId, promise);

        try {
            const user = await promise;
            this.userCache.set(userId, user);
            return user;
        } finally {
            this.pendingRequests.delete(userId);
        }
    }

    // Batch get users, utilizing cache
    async getBatchUsers(userIds) {
        // Filter out cached users
        const missingIds = userIds.filter(id => !this.userCache.has(id));
        
        if (missingIds.length > 0) {
            try {
                const users = await api.users.batchGet(missingIds);
                if (Array.isArray(users)) {
                    // Update cache with new users
                    users.forEach(user => {
                        this.userCache.set(user.id, user);
                    });
                }
            } catch (error) {
                console.error('Error fetching batch users:', error);
            }
        }

        // Return all requested users from cache
        return userIds.map(id => this.userCache.get(id)).filter(Boolean);
    }

    // Get user's nickname with fallback
    getNickname(userId) {
        const user = this.userCache.get(userId);
        return user?.nickname || 'Unknown User';
    }

    // Get user's avatar URL with fallback
    getAvatarUrl(userId) {
        const user = this.userCache.get(userId);
        return user?.avatar_url || null;
    }

    // Update cache with new user data
    updateUserCache(userData) {
        this.userCache.set(userData.id, userData);
    }

    // Update cache with multiple users
    updateBatchUserCache(users) {
        users.forEach(user => this.updateUserCache(user));
    }

    // Clear cache for specific user
    clearUserCache(userId) {
        this.userCache.delete(userId);
    }

    // Clear entire cache
    clearCache() {
        this.userCache.clear();
    }

    // Private method to fetch a single user
    async _fetchUser(userId) {
        try {
            const user = await api.users.get(userId);
            return user;
        } catch (error) {
            console.error(`Error fetching user ${userId}:`, error);
            throw error;
        }
    }
}

export const userService = new UserService();
