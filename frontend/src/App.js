import React from 'react';
import { ChakraProvider, Box } from '@chakra-ui/react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import ChatRoom from './components/ChatRoom';
import HomePage from './components/HomePage';

function App() {
  return (
    <ChakraProvider>
      <Router>
        <Box minH="100vh" bg="gray.100">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/chat/:sessionId" element={<ChatRoom />} />
          </Routes>
        </Box>
      </Router>
    </ChakraProvider>
  );
}

export default App;
