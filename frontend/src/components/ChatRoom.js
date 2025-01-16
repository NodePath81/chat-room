import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, VStack, Input, Button, Text, useToast, Flex } from '@chakra-ui/react';
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
    const [users, setUsers] = useState({});

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
                // Extract unique user IDs from history messages
                const userIds = [...new Set(historyMessages.map(msg => msg.userId))];
                // Fetch usernames for all users
                userIds.forEach(userId => fetchUsername(userId));
            });

            WebSocketService.onMessage((message) => {
                setMessages(prev => [...prev, message]);
                // Fetch username if not already known
                if (!users[message.userId]) {
                    fetchUsername(message.userId);
                }
            });
        }

        checkAndJoinSession();

        return () => {
            WebSocketService.disconnect();
            setMessages([]);
        };
    }, [sessionId, navigate, toast]);

    const fetchUsername = async (userId) => {
        try {
            const response = await fetch(`http://localhost:8080/api/users/${userId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            if (response.ok) {
                const userData = await response.json();
                setUsers(prev => ({
                    ...prev,
                    [userId]: userData.username
                }));
            }
        } catch (error) {
            console.error('Error fetching username:', error);
        }
    };

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
                        <Box 
                            key={index}
                            mb={2}
                            p={2}
                            bg="gray.50"
                            borderRadius="md"
                        >
                            <Text fontWeight="bold" color="blue.600">
                                {users[msg.userId] || 'Loading...'}
                            </Text>
                            <Text>{msg.content}</Text>
                        </Box>
                    ))}
                </Box>
                <Flex>
                    <Input
                        flex={1}
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
                        colorScheme="blue"
                    >
                        Send
                    </Button>
                </Flex>
            </VStack>
        </Box>
    );
}

export default ChatRoom; 