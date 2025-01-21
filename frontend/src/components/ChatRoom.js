import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import TopBar from './chat/TopBar';
import ChatBoard from './chat/ChatBoard';
import SendBar from './chat/SendBar';
import SessionService from '../services/session';
import { chatService } from '../services/chat';

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
            // Check user role and session details
            const role = await SessionService.checkSessionRole(currentSessionId);
            if (role === null) {
                navigate('/');
                return;
            }
            
            setIsCreator(role === 'creator');
            
            // Get session details
            const sessionDetails = await SessionService.getSession(currentSessionId);
            setSessionName(sessionDetails.name);
            console.log('Session details loaded:', sessionDetails);
            
            
            
            // Connect WebSocket
            try {
                await chatService.connectToSession(currentSessionId);
                console.log('WebSocket connection initiated');

                // Load initial messages
                const timestamp = new Date().getTime();
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
            const response = await chatService.getMessages(beforeTimestamp);
            const newMessages = response.messages || [];
            
            if (newMessages.length === 0) {
                setHasMore(false);
            return;
        }

            setMessages(prev => 
                beforeTimestamp ? [...newMessages, ...prev] : newMessages
            );
            setHasMore(response.hasMore);

            // Store the oldest message's timestamp
            if (newMessages.length > 0) {
                oldestTimestampRef.current = response.oldestTimestamp;
            }

            // Fetch usernames for all messages
            const userIds = new Set(newMessages.map(msg => msg.user_id));
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
        if (!messageData.trim()) return;
        
        try {
            await chatService.sendTextMessage(messageData);
        } catch (error) {
            console.error('Error sending message:', error);
        }
    };

    const handleImageUpload = async (file) => {
        try {
            await chatService.uploadImage(file);
        } catch (error) {
            console.error('Error uploading image:', error);
        }
    };

    const handleSettingsClick = () => {
        navigate(`/sessions/${currentSessionId}/manage`);
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-screen bg-gray-50">
                <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
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
            
            <SendBar
                onSendMessage={handleSendMessage}
                onImageUpload={handleImageUpload}
            />
        </div>
    );
}

export default ChatRoom; 