import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import LoginPage from './components/LoginPage';
import RegisterPage from './components/RegisterPage';
import HomePage from './components/HomePage';
import ChatRoom from './components/ChatRoom';
import UserPage from './components/UserPage';
import SessionManagePage from './components/SessionManagePage';
import SharePage from './components/SharePage';

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/" element={<HomePage />} />
          <Route path="/chat/:sessionId" element={<ChatRoom />} />
          <Route path="/profile" element={<UserPage />} />
          <Route path="/sessions/:id/manage" element={<SessionManagePage />} />
          <Route path="/share" element={<SharePage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
