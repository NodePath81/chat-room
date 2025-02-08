import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import TopBar from '../components/chat/TopBar';
import ChatBoard from '../components/chat/ChatBoard';
import SendBar from '../components/chat/SendBar';
import sessionService from '../services/session';
import { websocketService } from '../services/websocket';
import { userService } from '../services/user';

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
            // Initialize session and get details
            const sessionData = await sessionService.initializeSession(currentSessionId);
            setIsCreator(sessionData.isCreator);
            setSessionName(sessionData.session.name);

            // Connect WebSocket
            await websocketService.connect(currentSessionId);

            // Load initial messages
            const timestamp = new Date().toISOString();
            await loadMessages(timestamp);

            // Set up WebSocket message handler
            websocketService.onMessage((message) => {
                setMessages(prev => [...prev, message]);
                if (message.user_id) {
                    fetchMissingUsers(new Set([message.user_id]));
                }
            });

            // Load initial users
            const sessionUsers = await sessionService.getSessionUsers(currentSessionId);
            if (sessionUsers && sessionUsers.length > 0) {
                userService.updateBatchUserCache(sessionUsers);
                const newUsers = {};
                sessionUsers.forEach(user => {
                    if (user && user.id) {
                        newUsers[user.id] = user;
                    }
                });
                setUsers(prev => ({...prev, ...newUsers}));
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
            websocketService.disconnect();
        };
    }, [currentSessionId, initializeChat]);

    async function loadMessages(beforeTimestamp = null) {
        try {
            const response = await sessionService.getMessages(currentSessionId, beforeTimestamp);
            const newMessages = response.messages || [];
            
            if (newMessages.length === 0) {
                setHasMore(false);
                return;
            }

            setMessages(prev => {
                if (beforeTimestamp) {
                    return [...newMessages, ...prev];
                } else {
                    return newMessages;
                }
            });
            setHasMore(response.hasMore);

            // Store the oldest message's timestamp
            if (newMessages.length > 0) {
                oldestTimestampRef.current = newMessages[0].timestamp;
            }

            // Fetch usernames for new messages
            const userIds = new Set(newMessages.map(msg => msg.user_id));
            await fetchMissingUsers(userIds);
        } catch (error) {
            console.error('Error loading messages:', error);
        }
    }

    async function fetchMissingUsers(userIds) {
        const missingUserIds = Array.from(userIds).filter(id => !users[id]);
        if (missingUserIds.length === 0) return;

        try {
            const fetchedUsers = await userService.getBatchUsers(missingUserIds);
            const newUsers = {};
            fetchedUsers.forEach(user => {
                if (user && user.id) {
                    newUsers[user.id] = user;
                }
            });
            
            if (Object.keys(newUsers).length > 0) {
                setUsers(prev => ({...prev, ...newUsers}));
            }
        } catch (error) {
            console.error('Error fetching users:', error);
        }
    }

    // Add effect to monitor users state changes
    useEffect(() => {
        console.debug('Users state updated:', users);
    }, [users]);

    // Add effect to monitor messages and ensure we have all user data
    useEffect(() => {
        if (messages.length > 0) {
            const userIds = new Set(messages.map(msg => msg.user_id));
            fetchMissingUsers(userIds);
        }
    }, [messages]);

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
        try {
            websocketService.sendMessage({
                type: messageData.type,
                content: messageData.content
            });
        } catch (error) {
            console.error('Error sending message:', error);
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
            
            <SendBar onSendMessage={handleSendMessage} />
        </div>
    );
}

export default ChatRoom; 