import React, { createContext, useState, useContext, useEffect } from 'react';
import { authService } from '../services/auth';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(authService.getUser());
    const [isAuthenticated, setIsAuthenticated] = useState(!!authService.getToken());

    useEffect(() => {
        // Update authentication state when token changes
        const token = authService.getToken();
        setIsAuthenticated(!!token);
        setUser(authService.getUser());
    }, []);

    const login = (token, userData) => {
        authService.setToken(token);
        authService.setUser(userData);
        setUser(userData);
        setIsAuthenticated(true);
    };

    const logout = () => {
        authService.clearToken();
        authService.clearUser();
        setUser(null);
        setIsAuthenticated(false);
    };

    const value = {
        user,
        setUser,
        isAuthenticated,
        setIsAuthenticated,
        login,
        logout
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === null) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export default AuthContext; 