import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Box, VStack, Button, Text, useToast, Input } from '@chakra-ui/react';
import SessionService from '../services/session';

function HomePage() {
    const [sessions, setSessions] = useState([]);
    const [newSessionName, setNewSessionName] = useState('');
    const navigate = useNavigate();
    const toast = useToast();

    useEffect(() => {
        fetchSessions();
    }, []);

    const fetchSessions = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/sessions', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            if (!response.ok) {
                throw new Error('Failed to fetch sessions');
            }
            const data = await response.json();
            console.log('Fetched sessions:', data);
            
            // Map the response to match frontend expectations
            const mappedSessions = data.map(session => ({
                id: session.ID,
                name: session.name,
                users: session.Users?.map(user => ({
                    id: user.ID,
                    username: user.Username
                })) || []
            }));
            
            setSessions(mappedSessions);
        } catch (error) {
            console.error('Fetch error:', error);
            toast({
                title: "Error",
                description: "Failed to fetch sessions",
                status: "error",
                duration: 3000,
            });
        }
    };

    const createSession = async () => {
        if (!newSessionName.trim()) return;

        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/api/sessions', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name: newSessionName })
            });

            if (!response.ok) {
                const errorText = await response.text();
                console.error('Create session error:', errorText);
                throw new Error('Failed to create session');
            }

            const newSession = await response.json();
            console.log('Created session:', newSession);
            
            setNewSessionName('');
            await fetchSessions();
            
            if (newSession && newSession.ID) {
                await handleJoinSession(newSession.ID);
            } else {
                console.error('Invalid session data:', newSession);
                throw new Error('Invalid session ID');
            }
        } catch (error) {
            console.error('Creation error:', error);
            toast({
                title: "Error",
                description: "Failed to create session",
                status: "error",
                duration: 3000,
            });
        }
    };

    const handleJoinSession = async (sessionId) => {
        try {
            // Check if already a member
            const isMember = await SessionService.checkSessionMembership(sessionId);
            
            if (!isMember) {
                // Try to join
                const joined = await SessionService.joinSession(sessionId);
                if (!joined) {
                    toast({
                        title: "Error",
                        description: "Failed to join session",
                        status: "error",
                        duration: 3000,
                    });
                    return;
                }
                toast({
                    title: "Success",
                    description: "Successfully joined session",
                    status: "success",
                    duration: 3000,
                });
            }
            
            // Navigate to chat room after successful join
            navigate(`/chat/${sessionId}`);
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to join session",
                status: "error",
                duration: 3000,
            });
        }
    };

    return (
        <Box p={4}>
            <VStack spacing={4} align="stretch">
                <Box>
                    <Input
                        value={newSessionName}
                        onChange={(e) => setNewSessionName(e.target.value)}
                        placeholder="New session name"
                    />
                    <Button onClick={createSession} ml={2}>
                        Create Session
                    </Button>
                </Box>
                
                <Box>
                    <Text fontSize="xl" mb={4}>Available Sessions</Text>
                    <VStack spacing={2} align="stretch">
                        {sessions && sessions.length > 0 ? (
                            sessions.map(session => (
                                <Box 
                                    key={`session-${session.id}`}
                                    p={4} 
                                    borderWidth={1} 
                                    borderRadius="md"
                                    display="flex"
                                    justifyContent="space-between"
                                    alignItems="center"
                                >
                                    <Text>{session.name}</Text>
                                    <Button 
                                        onClick={() => session.id && handleJoinSession(session.id)}
                                        colorScheme="blue"
                                        isDisabled={!session.id}
                                    >
                                        Join Chat
                                    </Button>
                                </Box>
                            ))
                        ) : (
                            <Text>No sessions available</Text>
                        )}
                    </VStack>
                </Box>
            </VStack>
        </Box>
    );
}

export default HomePage; 