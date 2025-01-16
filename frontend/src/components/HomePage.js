import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import SessionService from '../services/session';
import { authService } from '../services/auth';

function HomePage() {
    const [sessions, setSessions] = useState([]);
    const [newSessionName, setNewSessionName] = useState('');
    const [user, setUser] = useState(authService.getUser());
    const navigate = useNavigate();

    useEffect(() => {
        fetchSessions();
        // Fetch latest user data
        const fetchUserData = async () => {
            const userData = await authService.fetchUserData();
            if (userData) {
                setUser(userData);
            }
        };
        fetchUserData();
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
                throw new Error('Failed to create session');
            }

            const newSession = await response.json();
            setNewSessionName('');
            await fetchSessions();
            
            if (newSession && newSession.ID) {
                await handleJoinSession(newSession.ID);
            }
        } catch (error) {
            console.error('Creation error:', error);
        }
    };

    const handleJoinSession = async (sessionId) => {
        try {
            const isMember = await SessionService.checkSessionMembership(sessionId);
            
            if (!isMember) {
                const joined = await SessionService.joinSession(sessionId);
                if (!joined) return;
            }
            
            navigate(`/chat/${sessionId}`);
        } catch (error) {
            console.error('Join error:', error);
        }
    };

    const handleLogout = () => {
        authService.logout();
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-gray-100 p-6">
            <div className="max-w-4xl mx-auto">
                <div className="flex justify-between items-center mb-8">
                    <h1 className="text-3xl font-bold text-gray-800">Chat Rooms</h1>
                    <div className="flex items-center space-x-4">
                        <button
                            onClick={() => navigate('/profile')}
                            className="flex items-center space-x-2 px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                        >
                            {user?.avatarUrl ? (
                                <img
                                    src={user.avatarUrl}
                                    alt="Profile"
                                    className="w-6 h-6 rounded-full object-cover"
                                />
                            ) : (
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                </svg>
                            )}
                            <span>Profile</span>
                        </button>
                        <button
                            onClick={handleLogout}
                            className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                        >
                            Logout
                        </button>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6 mb-6">
                    <div className="flex gap-4">
                        <input
                            type="text"
                            value={newSessionName}
                            onChange={(e) => setNewSessionName(e.target.value)}
                            placeholder="New session name"
                            className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <button
                            onClick={createSession}
                            className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                        >
                            Create Session
                        </button>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow-md p-6">
                    <h2 className="text-xl font-semibold mb-4">Available Sessions</h2>
                    <div className="space-y-3">
                        {sessions.length > 0 ? (
                            sessions.map(session => (
                                <div
                                    key={`session-${session.id}`}
                                    className="flex justify-between items-center p-4 bg-gray-50 rounded-md hover:bg-gray-100"
                                >
                                    <span className="text-gray-700">{session.name}</span>
                                    <button
                                        onClick={() => session.id && handleJoinSession(session.id)}
                                        className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                        disabled={!session.id}
                                    >
                                        Join Chat
                                    </button>
                                </div>
                            ))
                        ) : (
                            <p className="text-gray-500 text-center py-4">No sessions available</p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

export default HomePage; 