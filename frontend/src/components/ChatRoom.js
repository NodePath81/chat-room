import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import WebSocketService from '../services/websocket';
import SessionService from '../services/session';

function ChatRoom() {
    const { sessionId } = useParams();
    const navigate = useNavigate();
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [isJoined, setIsJoined] = useState(false);
    const [isLoading, setIsLoading] = useState(true);
    const [users, setUsers] = useState({});

    useEffect(() => {
        const fetchUsername = async (userId) => {
            if (users[userId]) return;
            
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

        async function checkAndJoinSession() {
            setIsLoading(true);
            const isMember = await SessionService.checkSessionMembership(sessionId);
            
            if (!isMember) {
                navigate('/');
                return;
            }
            
            setIsJoined(true);
            setIsLoading(false);
            
            WebSocketService.connect(parseInt(sessionId, 10));
            
            WebSocketService.onHistory((historyMessages) => {
                setMessages(historyMessages);
                const userIds = [...new Set(historyMessages.map(msg => msg.userId))];
                userIds.forEach(userId => fetchUsername(userId));
            });

            WebSocketService.onMessage((message) => {
                setMessages(prev => [...prev, message]);
                if (message.userId) {
                    fetchUsername(message.userId);
                }
            });
        }

        checkAndJoinSession();

        return () => {
            WebSocketService.disconnect();
            setMessages([]);
        };
    }, [sessionId, navigate]);

    const handleSend = () => {
        if (!isJoined || !newMessage.trim()) return;
        WebSocketService.sendMessage(newMessage, parseInt(sessionId, 10));
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
        <div className="min-h-screen bg-gray-100 p-4">
            <div className="max-w-4xl mx-auto bg-white rounded-lg shadow-md h-[calc(100vh-2rem)] flex flex-col">
                <div className="p-4 border-b bg-gray-50">
                    <h1 className="text-xl font-semibold text-gray-800">Chat Room</h1>
                </div>
                
                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                    {messages.map((msg, index) => (
                        <div 
                            key={index}
                            className="bg-gray-50 rounded-lg p-3 shadow-sm hover:shadow-md transition-shadow"
                        >
                            <div className="font-semibold text-blue-600">
                                {users[msg.userId] || 'Loading...'}
                            </div>
                            <div className="text-gray-700 mt-1">
                                {msg.content}
                            </div>
                        </div>
                    ))}
                </div>
                
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