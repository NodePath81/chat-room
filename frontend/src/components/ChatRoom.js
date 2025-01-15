import React, { useEffect, useState } from 'react';
import { Box, VStack } from '@chakra-ui/react';
import { useParams } from 'react-router-dom';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import { WebSocketService } from '../services/websocket';
import { authService } from '../services/auth';

const ws = new WebSocketService(process.env.REACT_APP_WS_URL || 'ws://localhost:8080');

function ChatRoom() {
  const { sessionId } = useParams();
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    if (sessionId) {
      const token = authService.getToken();
      ws.connect(sessionId, token);
    }

    return () => {
      ws.disconnect();
    };
  }, [sessionId]);

  const handleSendMessage = (content) => {
    ws.sendMessage(content);
  };

  return (
    <Box h="100vh" p={4}>
      <VStack h="full" spacing={4}>
        <MessageList messages={messages} />
        <MessageInput onSendMessage={handleSendMessage} />
      </VStack>
    </Box>
  );
}

export default ChatRoom; 