import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import WebSocketService from '../services/websocket';
import SessionService from '../services/session';
import { API_ENDPOINTS } from '../services/api';
import { authService } from '../services/auth';

function ChatRoom() {
    const { sessionId } = useParams();
    const navigate = useNavigate();
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [isJoined, setIsJoined] = useState(false);
    const [isLoading, setIsLoading] = useState(true);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [users, setUsers] = useState({});
    const [userRole, setUserRole] = useState(null);
    const messagesEndRef = useRef(null);
    const messageListRef = useRef(null);
    const [showUpdateZone, setShowUpdateZone] = useState(false);
    const updateZoneRef = useRef(null);
    const [updateZoneExpanded, setUpdateZoneExpanded] = useState(false);
    const lastScrollTopRef = useRef(0);
    const loadingRef = useRef(false);
    const lastLoadedMessageRef = useRef(null);
    const lastRequestTimeRef = useRef(0);  // Track time of last request

    const scrollToBottom = () => {
        if (messageListRef.current) {
            messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
        }
    };

    const canMakeRequest = () => {
        const now = Date.now();
        const timeSinceLastRequest = now - lastRequestTimeRef.current;
        return timeSinceLastRequest >= 500;
    };

    const positionOldestMessage = (messageId) => {
        if (messageListRef.current) {
            const oldestVisibleMessage = document.querySelector(`[data-message-id="${messageId}"]`);
            if (oldestVisibleMessage) {
                oldestVisibleMessage.scrollIntoView({ block: 'start', behavior: 'auto' });
            }
        }
    };

    const fetchUsername = async (userId) => {
        if (users[userId]) return;
        
        try {
            const response = await fetch(API_ENDPOINTS.USERS.GET(userId), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });
            if (response.ok) {
                const userData = await response.json();
                setUsers(prev => ({
                    ...prev,
                    [userId]: {
                        nickname: userData.nickname,
                        avatarUrl: userData.avatarUrl
                    }
                }));
            }
        } catch (error) {
            console.error('Error fetching user data:', error);
        }
    };

    const loadMessages = async (beforeId = null) => {
        if (loadingRef.current) return [];
        if (!canMakeRequest()) return [];
        
        try {
            lastRequestTimeRef.current = Date.now();
            loadingRef.current = true;
            setIsLoadingMore(true);
            
            const response = await fetch(API_ENDPOINTS.SESSIONS.GET_MESSAGES(sessionId, {
                before: beforeId,
                limit: 5
            }), {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to fetch messages');
            }

            const data = await response.json();
            const userIds = [...new Set(data.messages.map(msg => msg.userId))];
            userIds.forEach(userId => fetchUsername(userId));

            if (beforeId) {
                lastLoadedMessageRef.current = beforeId;
                setMessages(prev => [...data.messages.reverse(), ...prev]);
            } else {
                setMessages(data.messages.reverse());
            }
            setHasMore(data.hasMore);
            
            return data.messages;
        } catch (error) {
            console.error('Error loading messages:', error);
            return [];
        } finally {
            loadingRef.current = false;
            setIsLoadingMore(false);
        }
    };

    // Effect to handle scroll positioning after messages update
    useEffect(() => {
        if (lastLoadedMessageRef.current && messages.length > 0) {
            positionOldestMessage(lastLoadedMessageRef.current);
            lastLoadedMessageRef.current = null;
        }
    }, [messages]);

    const handleScroll = () => {
        if (!messageListRef.current || loadingRef.current) return;
        
        const { scrollTop } = messageListRef.current;

        if (scrollTop < 20) {
            if (!showUpdateZone) {
                setShowUpdateZone(true);
                setUpdateZoneExpanded(false);
            } else if (!updateZoneExpanded && scrollTop < lastScrollTopRef.current) {
                setUpdateZoneExpanded(true);
            } else if (updateZoneExpanded && scrollTop < lastScrollTopRef.current) {
                if (hasMore && !loadingRef.current && canMakeRequest()) {
                    const oldestMessage = messages[0];
                    if (oldestMessage) {
                        loadMessages(oldestMessage.id);
                    }
                }
            }
        } else if (scrollTop > 50) {
            setShowUpdateZone(false);
            setUpdateZoneExpanded(false);
        }

        lastScrollTopRef.current = scrollTop;
    };

    // Initialize lastRequestTimeRef in component mount
    useEffect(() => {
        console.log('Initializing request timer');
        lastRequestTimeRef.current = 0;  // Explicitly initialize to 0
        
        return () => {
            console.log('Cleaning up refs');
            lastScrollTopRef.current = 0;
            lastRequestTimeRef.current = 0;
            loadingRef.current = false;
        };
    }, []);

    useEffect(() => {
        const currentSessionId = parseInt(sessionId, 10);

        const checkUserRole = async () => {
            try {
                const response = await fetch(API_ENDPOINTS.SESSIONS.CHECK_ROLE(currentSessionId), {
                    headers: {
                        'Authorization': `Bearer ${authService.getToken()}`
                    }
                });
                if (response.ok) {
                    const data = await response.json();
                    setUserRole(data.role);
                }
            } catch (error) {
                console.error('Error checking user role:', error);
            }
        };

        async function checkAndJoinSession() {
            setIsLoading(true);
            const isMember = await SessionService.checkSessionMembership(currentSessionId);
            
            if (!isMember) {
                navigate('/');
                return;
            }
            
            await checkUserRole();
            setIsJoined(true);
            
            // Load initial messages
            await loadMessages();

            // Ensure we're at the bottom after initial load
            setTimeout(() => {
                if (messageListRef.current) {
                    messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
                    lastScrollTopRef.current = messageListRef.current.scrollTop;
                }
            }, 100);
            
            WebSocketService.connect(currentSessionId);
            
            WebSocketService.onMessage((message) => {
                // Store the current scroll position and the height before adding new message
                const scrollPos = messageListRef.current?.scrollTop || 0;
                const oldHeight = messageListRef.current?.scrollHeight || 0;

                setMessages(prev => [...prev, message]);
                if (message.userId) {
                    fetchUsername(message.userId);
                }

                // After state update, adjust scroll position if user was not at bottom
                requestAnimationFrame(() => {
                    if (messageListRef.current) {
                        const newHeight = messageListRef.current.scrollHeight;
                        const isAtBottom = (scrollPos + messageListRef.current.clientHeight + 100) >= oldHeight;
                        
                        if (isAtBottom) {
                            scrollToBottom();
                        } else {
                            // Maintain relative scroll position
                            messageListRef.current.scrollTop = scrollPos + (newHeight - oldHeight);
                        }
                    }
                });
            }, currentSessionId);

            setIsLoading(false);
        }

        checkAndJoinSession();

        return () => {
            WebSocketService.removeHandlers(currentSessionId);
            WebSocketService.disconnect(currentSessionId);
            setMessages([]);
            setUsers({});
            lastScrollTopRef.current = 0;
        };
    }, [sessionId, navigate]);

    const handleSend = () => {
        if (!isJoined || !newMessage.trim()) return;
        
        const currentSessionId = parseInt(sessionId, 10);
        WebSocketService.sendMessage(newMessage, currentSessionId);
        setNewMessage('');
    };

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
                <div className="w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
                <div className="text-xl text-gray-600 font-semibold">Connecting to chat...</div>
                <div className="text-sm text-gray-500 mt-2">Please wait while we set things up</div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-screen bg-gray-100">
            <div className="flex-1 flex flex-col max-w-4xl mx-auto w-full bg-white shadow-lg">
                {/* Top Bar */}
                <div className="p-4 border-b bg-gray-50 flex justify-between items-center">
                    <h1 className="text-xl font-semibold text-gray-800">Chat Room</h1>
                    {userRole === 'creator' && (
                        <button
                            onClick={() => navigate(`/sessions/${sessionId}/manage`)}
                            className="text-gray-600 hover:text-gray-800 focus:outline-none"
                            aria-label="Settings"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" viewBox="0 0 20 20" fill="currentColor">
                                <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
                            </svg>
                        </button>
                    )}
                </div>

                {/* Message List Container */}
                <div 
                    className="flex-1 overflow-hidden relative"
                    style={{
                        height: 'calc(100vh - 200px)',
                    }}
                >
                    {isLoading ? (
                        <div className="absolute inset-0 flex items-center justify-center">
                            <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                        </div>
                    ) : (
                        <div 
                            ref={messageListRef}
                            onScroll={handleScroll}
                            className="absolute inset-0 overflow-y-auto px-4 py-2 space-y-4"
                        >
                            <div 
                                ref={updateZoneRef}
                                className={`sticky top-0 left-0 right-0 transition-all duration-300 overflow-hidden ${
                                    showUpdateZone ? 'mb-4' : 'mb-0'
                                } ${
                                    updateZoneExpanded ? 'h-16 opacity-100' : 'h-0 opacity-0'
                                }`}
                                style={{
                                    transform: updateZoneExpanded ? 'translateY(0)' : 'translateY(-100%)'
                                }}
                            >
                                <div className="flex items-center justify-center h-full bg-blue-50 rounded-lg">
                                    {isLoadingMore ? (
                                        <div className="flex items-center space-x-2">
                                            <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                                            <span className="text-blue-600">Loading more messages...</span>
                                        </div>
                                    ) : hasMore ? (
                                        <span className="text-blue-600">
                                            {updateZoneExpanded ? 'Loading more messages...' : 'Scroll up to load more'}
                                        </span>
                                    ) : (
                                        <span className="text-gray-500">No more messages</span>
                                    )}
                                </div>
                            </div>

                            <div className="space-y-4 min-h-full">
                                {messages.map((msg, index) => (
                                    msg && msg.content && (
                                        <div 
                                            key={msg.id || index}
                                            data-message-id={msg.id}
                                            className="flex items-start gap-3 hover:bg-gray-50 p-2 rounded-lg transition-colors"
                                        >
                                            <div className="flex-shrink-0 w-10 h-10">
                                                {users[msg.userId]?.avatarUrl ? (
                                                    <img
                                                        src={users[msg.userId].avatarUrl}
                                                        alt={`${users[msg.userId]?.nickname || 'User'}'s avatar`}
                                                        className="w-10 h-10 rounded-full object-cover"
                                                    />
                                                ) : (
                                                    <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                                                        <span className="text-blue-600 font-semibold text-lg">
                                                            {(users[msg.userId]?.nickname || 'U')[0].toUpperCase()}
                                                        </span>
                                                    </div>
                                                )}
                                            </div>
                                            <div className="flex-1">
                                                <div className="flex justify-between items-center">
                                                    <div className="font-semibold text-blue-600">
                                                        {users[msg.userId]?.nickname || 'Loading...'}
                                                    </div>
                                                    <div className="text-xs text-gray-500">
                                                        {new Date(msg.timestamp).toLocaleString()}
                                                    </div>
                                                </div>
                                                <div className="text-gray-700 mt-1">
                                                    {msg.content}
                                                </div>
                                            </div>
                                        </div>
                                    )
                                ))}
                            </div>
                            <div ref={messagesEndRef} />
                        </div>
                    )}
                </div>

                {/* Send Message Bar */}
                <div className="border-t p-4 bg-gray-50">
                    <div className="flex gap-4">
                        <input
                            type="text"
                            value={newMessage}
                            onChange={(e) => setNewMessage(e.target.value)}
                            onKeyPress={(e) => e.key === 'Enter' && handleSend()}
                            placeholder="Type a message..."
                            className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                            disabled={!isJoined}
                        />
                        <button
                            onClick={handleSend}
                            disabled={!isJoined}
                            className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                        >
                            Send
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default ChatRoom; 