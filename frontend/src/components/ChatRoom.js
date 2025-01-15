import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, VStack, Input, Button, Text, useToast } from '@chakra-ui/react';
import WebSocketService from '../services/websocket';
import SessionService from '../services/session';

function ChatRoom() {
    const { sessionId } = useParams();
    const navigate = useNavigate();
    const toast = useToast();
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [isJoined, setIsJoined] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        async function checkAndJoinSession() {
            setIsLoading(true);
            const isMember = await SessionService.checkSessionMembership(sessionId);
            
            if (!isMember) {
                toast({
                    title: "Error",
                    description: "You are not a member of this session",
                    status: "error",
                    duration: 3000,
                });
                navigate('/');
                return;
            }
            
            setIsJoined(true);
            setIsLoading(false);
            
            // Convert sessionId to number when connecting
            WebSocketService.connect(parseInt(sessionId, 10));
            
            WebSocketService.onHistory((historyMessages) => {
                setMessages(historyMessages);
            });

            WebSocketService.onMessage((message) => {
                setMessages(prev => [...prev, message]);
            });
        }

        checkAndJoinSession();

        return () => {
            WebSocketService.disconnect();
            setMessages([]);
        };
    }, [sessionId, navigate, toast]);

    const handleSend = () => {
        if (!isJoined) {
            toast({
                title: "Cannot send message",
                description: "You must join the session first",
                status: "warning",
                duration: 3000,
            });
            return;
        }

        if (newMessage.trim()) {
            // Convert sessionId to number when sending
            WebSocketService.sendMessage(newMessage, parseInt(sessionId, 10));
            setNewMessage('');
        }
    };

    if (isLoading) {
        return (
            <Box p={4} textAlign="center">
                Loading...
            </Box>
        );
    }

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
                        isDisabled={!isJoined}
                    />
                    <Button 
                        onClick={handleSend} 
                        ml={2}
                        isDisabled={!isJoined}
                    >
                        Send
                    </Button>
                </Box>
            </VStack>
        </Box>
    );
}

export default ChatRoom; 