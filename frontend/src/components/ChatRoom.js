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
    const [selectedImage, setSelectedImage] = useState(null);
    const [imagePreview, setImagePreview] = useState(null);
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
                        avatarUrl: userData.avatar_url
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
            const userIds = [...new Set(data.messages.map(msg => msg.user_id))];
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
        // Remove parseInt as sessionId should remain a string for UUID
        const currentSessionId = sessionId;

        async function initializeChat() {
            setIsLoading(true);
            try {
                // Check user role, which also verifies session membership
                const role = await SessionService.checkSessionRole(currentSessionId);
                if (role === null) {
                    // User is not a member or session doesn't exist
                    navigate('/');
                    return;
                }
                
                setUserRole(role);
                setIsJoined(true);
                
                // Load initial messages through API
                await loadMessages();

                // Ensure we're at the bottom after initial load
                setTimeout(() => {
                    if (messageListRef.current) {
                        messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
                        lastScrollTopRef.current = messageListRef.current.scrollTop;
                    }
                }, 100);
                
                // Connect WebSocket after confirming membership
                WebSocketService.connect(currentSessionId);
                
                WebSocketService.onMessage((message) => {
                    // Store the current scroll position and the height before adding new message
                    const scrollPos = messageListRef.current?.scrollTop || 0;
                    const oldHeight = messageListRef.current?.scrollHeight || 0;

                    setMessages(prev => [...prev, message]);
                    if (message.user_id) {
                        fetchUsername(message.user_id);
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
            } catch (error) {
                console.error('Error initializing chat:', error);
                navigate('/');
            } finally {
                setIsLoading(false);
            }
        }

        initializeChat();

        return () => {
            WebSocketService.removeHandlers(currentSessionId);
            WebSocketService.disconnect(currentSessionId);
            setMessages([]);
            setUsers({});
            lastScrollTopRef.current = 0;
        };
    }, [sessionId, navigate]);

    const handleImageUpload = async (event) => {
        const file = event.target.files[0];
        if (!file) return;

        // Check file type
        if (!file.type.startsWith('image/')) {
            alert('Please select an image file');
            return;
        }

        // Check file size (max 5MB)
        if (file.size > 5 * 1024 * 1024) {
            alert('Image size should be less than 5MB');
            return;
        }

        setSelectedImage(file);
        
        // Create preview URL
        const previewUrl = URL.createObjectURL(file);
        setImagePreview(previewUrl);
    };

    const clearImageSelection = () => {
        if (imagePreview) {
            URL.revokeObjectURL(imagePreview);
        }
        setSelectedImage(null);
        setImagePreview(null);
    };

    const handleSend = async (imageFile = null) => {
        if ((!newMessage && !selectedImage) || !isJoined) {
            console.log('Message send blocked:', { hasNewMessage: !!newMessage, hasSelectedImage: !!selectedImage, isJoined });
            return;
        }

        try {
            if (selectedImage) {
                console.log('Uploading image...');
                // Upload image first
                const formData = new FormData();
                formData.append('image', selectedImage);

                const response = await fetch(API_ENDPOINTS.SESSIONS.UPLOAD_MESSAGE_IMAGE(sessionId), {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${authService.getToken()}`
                    },
                    body: formData
                });

                if (!response.ok) {
                    throw new Error('Failed to upload image');
                }

                const data = await response.json();
                console.log('Image uploaded successfully:', data);
                
                // No need to send via WebSocket - the API handler will broadcast
                clearImageSelection();
            } else if (newMessage.trim()) {
                console.log('Sending text message:', newMessage.trim());
                // Send text message via WebSocket
                WebSocketService.sendMessage({
                    type: 'text',
                    content: newMessage.trim(),
                    sessionId: sessionId
                });
                setNewMessage('');
            }
        } catch (error) {
            console.error('Error sending message:', error);
            alert('Failed to send message. Please try again.');
        }
    };

    // Clean up preview URL on unmount
    useEffect(() => {
        return () => {
            if (imagePreview) {
                URL.revokeObjectURL(imagePreview);
            }
        };
    }, [imagePreview]);

    // Message type components
    const MessageContent = ({ message }) => {
        switch (message.type) {
            case 'image':
                return (
                    <div className="mt-2">
                        <img 
                            src={message.content} 
                            alt="Message attachment" 
                            className="max-w-sm rounded-lg shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                            onClick={() => window.open(message.content, '_blank')}
                            onError={(e) => {
                                e.target.onerror = null;
                                e.target.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%23999"><path d="M13 14h-2v-2h2m0-2h-2V7h2m-1-5C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/></svg>';
                                e.target.className = "w-16 h-16 opacity-50";
                                e.target.title = "Failed to load image";
                            }}
                        />
                    </div>
                );
            case 'text':
            default:
                return (
                    <div className="text-gray-700 mt-1 whitespace-pre-wrap break-words">
                        {message.content}
                    </div>
                );
        }
    };

    const MessageBubble = ({ message, user }) => (
        <div className="flex items-start gap-3 hover:bg-gray-50 p-2 rounded-lg transition-colors">
            <div className="flex-shrink-0 w-10 h-10">
                {user?.avatarUrl ? (
                    <img
                        src={user.avatarUrl}
                        alt={`${user?.nickname || 'User'}'s avatar`}
                        className="w-10 h-10 rounded-full object-cover"
                        onError={(e) => {
                            e.target.onerror = null;
                            e.target.src = `data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%23999"><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>`;
                        }}
                    />
                ) : (
                    <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                        <span className="text-blue-600 font-semibold text-lg">
                            {(user?.nickname || 'U')[0].toUpperCase()}
                        </span>
                    </div>
                )}
            </div>
            <div className="flex-1 min-w-0">
                <div className="flex justify-between items-center">
                    <div className="font-semibold text-blue-600 truncate">
                        {user?.nickname || 'Loading...'}
                    </div>
                    <div className="text-xs text-gray-500 flex-shrink-0 ml-2">
                        {new Date(message.timestamp).toLocaleString()}
                    </div>
                </div>
                <MessageContent message={message} />
            </div>
        </div>
    );

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
                                        <div key={msg.id || index} data-message-id={msg.id}>
                                            <MessageBubble 
                                                message={msg} 
                                                user={users[msg.user_id]} 
                                            />
                                        </div>
                                    )
                                ))}
                            </div>
                            <div ref={messagesEndRef} />
                        </div>
                    )}
                </div>

                {/* Send Message Bar */}
                <div className="border-t bg-gray-50">
                    {/* Image Preview */}
                    {imagePreview && (
                        <div className="p-4 border-b">
                            <div className="relative inline-block">
                                <img 
                                    src={imagePreview} 
                                    alt="Upload preview" 
                                    className="max-h-32 rounded-lg shadow-sm"
                                />
                                <button
                                    onClick={clearImageSelection}
                                    className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-1 hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                                >
                                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                        <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                                    </svg>
                                </button>
                            </div>
                        </div>
                    )}

                    {/* Message Input */}
                    <div className="p-4">
                        <div className="flex gap-4">
                            <div className="flex-1 flex gap-2">
                                <input
                                    type="text"
                                    value={newMessage}
                                    onChange={(e) => setNewMessage(e.target.value)}
                                    onKeyPress={(e) => e.key === 'Enter' && handleSend()}
                                    placeholder="Type a message..."
                                    className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                                    disabled={!isJoined}
                                />
                                <label className="p-2 bg-blue-100 rounded-md cursor-pointer hover:bg-blue-200 transition-colors">
                                    <input
                                        type="file"
                                        accept="image/*"
                                        onChange={handleImageUpload}
                                        className="hidden"
                                        disabled={!isJoined}
                                    />
                                    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                </label>
                            </div>
                            <button
                                onClick={handleSend}
                                disabled={!isJoined || (!newMessage && !selectedImage)}
                                className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                            >
                                Send
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default ChatRoom; 