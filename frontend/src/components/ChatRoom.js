import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Box, VStack, Input, Button, Text } from '@chakra-ui/react';
import WebSocketService from '../services/websocket';

function ChatRoom() {
    const { sessionId } = useParams();
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');

    useEffect(() => {
        WebSocketService.connect(sessionId);
        
        // Handle history messages
        WebSocketService.onHistory((historyMessages) => {
            setMessages(historyMessages);
        });

        // Handle new messages
        WebSocketService.onMessage((message) => {
            setMessages(prev => [...prev, message]);
        });

        return () => WebSocketService.disconnect();
    }, [sessionId]);

    const handleSend = () => {
        if (newMessage.trim()) {
            WebSocketService.sendMessage(newMessage, parseInt(sessionId));
            setNewMessage('');
        }
    };

    return (
        <Box p={4}>
            <VStack spacing={4} align="stretch" h="80vh">
                <Box flex={1} overflowY="auto" borderWidth={1} p={4}>
                    {messages.map((msg, index) => (
                        <Text key={index}>
                            {msg.content}
                        </Text>
                    ))}
                </Box>
                <Box>
                    <Input
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        placeholder="Type a message..."
                        onKeyPress={(e) => e.key === 'Enter' && handleSend()}
                    />
                    <Button onClick={handleSend} ml={2}>
                        Send
                    </Button>
                </Box>
            </VStack>
        </Box>
    );
}

export default ChatRoom; 