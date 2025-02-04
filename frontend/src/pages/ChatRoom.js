import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import TopBar from '../components/chat/TopBar';
import ChatBoard from '../components/chat/ChatBoard';
import SendBar from '../components/chat/SendBar';
import { api } from '../services/api';
import { chatService } from '../services/chat';
import SessionService from '../services/session';

function ChatRoom() {
    const navigate = useNavigate();
    const { sessionId: currentSessionId } = useParams();
    const [isLoading, setIsLoading] = useState(true);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [showUpdateZone, setShowUpdateZone] = useState(false);
    const [updateZoneExpanded, setUpdateZoneExpanded] = useState(false);
    const [messages, setMessages] = useState([]);
    const [users, setUsers] = useState({});
    const [sessionName, setSessionName] = useState('');
    const [isCreator, setIsCreator] = useState(false);

    const oldestTimestampRef = React.useRef(null);

    const initializeChat = useCallback(async () => {
        setIsLoading(true);
        try {
            // Get session token first
            const tokenResponse = await api.sessions.getToken(currentSessionId);
            if (!tokenResponse || !tokenResponse.token) {
                console.error('Failed to get session token');
                navigate('/');
                return;
            }

            // Check user role and session details
            const roleResponse = await api.sessions.checkRole(currentSessionId);
            if (!roleResponse || !roleResponse.role) {
                console.error('Failed to get user role');
                navigate('/');
                return;
            }
            
            setIsCreator(roleResponse.role === 'creator');
            
            // Get session details
            const sessionDetails = await api.sessions.get(currentSessionId);
            if (!sessionDetails) {
                console.error('Failed to get session details');
                navigate('/');
                return;
            }

            setSessionName(sessionDetails.name);
            console.log('Session details loaded:', sessionDetails);
            
            // Connect WebSocket
            try {
                await chatService.connectToSession(currentSessionId);
                console.log('WebSocket connection initiated');

                // Load initial messages
                const timestamp = new Date().toISOString();
                await loadMessages(timestamp);
                console.log('Initial messages loaded');
                
                chatService.onMessageReceived((message) => {
                    console.log('Received message:', message);
                    setMessages(prev => [...prev, message]);
                    if (message.user_id) {
                        fetchUsername(message.user_id);
                    }
                });
            } catch (error) {
                console.error('Error setting up WebSocket:', error);
                // Continue loading the chat even if WebSocket fails
            }
        } catch (error) {
            console.error('Error initializing chat:', error);
            navigate('/');
        } finally {
            setIsLoading(false);
        }
    }, [currentSessionId, navigate]);

    useEffect(() => {
        initializeChat();
        return () => {
            chatService.disconnectFromSession(currentSessionId);
        };
    }, [currentSessionId, initializeChat]);

    async function loadMessages(beforeTimestamp = null) {
        try {
            const response = await api.sessions.getMessages(currentSessionId, {
                before: beforeTimestamp,
                limit: 50
            });
            const newMessages = response.messages || [];
            
            if (newMessages.length === 0) {
                setHasMore(false);
                return;
            }

            // Sort messages by timestamp in ascending order (oldest first)
            const sortedMessages = [...newMessages].sort((a, b) => 
                new Date(a.timestamp) - new Date(b.timestamp)
            );

            setMessages(prev => {
                if (beforeTimestamp) {
                    // When loading older messages, put them before existing messages
                    return [...sortedMessages, ...prev];
                } else {
                    // For initial load, just use the sorted messages
                    return sortedMessages;
                }
            });
            setHasMore(response.has_more);

            // Store the oldest message's timestamp
            if (sortedMessages.length > 0) {
                oldestTimestampRef.current = sortedMessages[0].timestamp;
            }

            // Fetch usernames for all messages
            const userIds = new Set(sortedMessages.map(msg => msg.user_id));
            for (const userId of userIds) {
                await fetchUsername(userId);
            }
        } catch (error) {
            console.error('Error loading messages:', error);
        }
    }

    async function fetchUsername(userId) {
        if (!users[userId]) {
            try {
                const userData = await chatService.fetchUserData(userId);
                setUsers(prev => ({
                    ...prev,
                    [userId]: userData
                }));
            } catch (error) {
                console.error('Error fetching user data:', error);
            }
        }
    }

    const handleLoadMore = async () => {
        if (isLoadingMore || !hasMore) return;
        
        setIsLoadingMore(true);
        setUpdateZoneExpanded(true);
        
        try {
            await loadMessages(oldestTimestampRef.current);
        } finally {
            setIsLoadingMore(false);
            setUpdateZoneExpanded(false);
        }
    };

    const handleUpdateZoneChange = (show, expanded) => {
        setShowUpdateZone(show);
        setUpdateZoneExpanded(expanded);
    };

    const handleSendMessage = async (messageData) => {
        console.debug('ChatRoom: Handling message:', messageData);
        
        if (!messageData.content || !messageData.type) {
            console.error('ChatRoom: Invalid message format:', messageData);
            return;
        }
        
        try {
            if (messageData.type === 'image') {
                console.debug('ChatRoom: Processing image upload');
                const formData = new FormData();
                formData.append('image', messageData.content);
                await api.sessions.uploadMessageImage(currentSessionId, formData);
            } else {
                console.debug('ChatRoom: Sending text message');
                await chatService.sendTextMessage(messageData.content);
            }
        } catch (error) {
            console.error('ChatRoom: Error sending message:', error);
        }
    };

    const handleImageUpload = async (file) => {
        try {
            console.debug('ChatRoom: Starting image upload:', file.name);
            const response = await chatService.uploadImage(file);
            console.debug('ChatRoom: Image upload completed:', response);
        } catch (error) {
            console.error('ChatRoom: Error uploading image:', error);
            // TODO: Show error notification to user
        }
    };

    const handleSettingsClick = () => {
        navigate(`/sessions/${currentSessionId}/manage`);
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-screen bg-gray-50">
                <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-screen bg-gray-50">
            <TopBar
                sessionName={sessionName}
                isCreator={isCreator}
                onSettingsClick={handleSettingsClick}
            />
            
            <div className="flex-1 overflow-hidden">
                <div className="max-w-7xl mx-auto h-full">
                    <ChatBoard
                        messages={messages}
                        users={users}
                        isLoading={isLoading}
                        isLoadingMore={isLoadingMore}
                        hasMore={hasMore}
                        showUpdateZone={showUpdateZone}
                        updateZoneExpanded={updateZoneExpanded}
                        onScroll={handleLoadMore}
                        onUpdateZoneChange={handleUpdateZoneChange}
                    />
                </div>
            </div>
            
            <SendBar
                onSendMessage={handleSendMessage}
                onImageUpload={handleImageUpload}
            />
        </div>
    );
}

export default ChatRoom; 