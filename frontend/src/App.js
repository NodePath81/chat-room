import React from 'react';
import { ChakraProvider, Box } from '@chakra-ui/react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import ChatRoom from './components/ChatRoom';
import HomePage from './components/HomePage';
import LoginPage from './components/LoginPage';
import UserPage from './components/UserPage';
import { authService } from './services/auth';

function PrivateRoute({ children }) {
  return authService.isAuthenticated() ? children : <Navigate to="/login" />;
}

function App() {
  return (
    <ChakraProvider>
      <Router>
        <Box minH="100vh" bg="gray.100">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
              path="/"
              element={
                <PrivateRoute>
                  <HomePage />
                </PrivateRoute>
              }
            />
            <Route
              path="/chat/:sessionId"
              element={
                <PrivateRoute>
                  <ChatRoom />
                </PrivateRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <PrivateRoute>
                  <UserPage />
                </PrivateRoute>
              }
            />
          </Routes>
        </Box>
      </Router>
    </ChakraProvider>
  );
}

export default App;
