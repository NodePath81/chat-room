import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import PrivateRoute from './components/PrivateRoute';

// Page imports
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import UserPage from './pages/UserPage';
import ChatRoom from './pages/ChatRoom';
import SessionManagePage from './pages/SessionManagePage';
import SharePage from './pages/SharePage';

function App() {
    return (
        <AuthProvider>
            <Router>
                <Routes>
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/register" element={<RegisterPage />} />
                    <Route path="/" element={<PrivateRoute><HomePage /></PrivateRoute>} />
                    <Route path="/profile" element={<PrivateRoute><UserPage /></PrivateRoute>} />
                    <Route path="/sessions/:sessionId" element={<PrivateRoute><ChatRoom /></PrivateRoute>} />
                    <Route path="/sessions/:sessionId/manage" element={<PrivateRoute><SessionManagePage /></PrivateRoute>} />
                    <Route path="/share" element={<PrivateRoute><SharePage /></PrivateRoute>} />
                    <Route path="*" element={<Navigate to="/" />} />
                </Routes>
            </Router>
        </AuthProvider>
    );
}

export default App;
