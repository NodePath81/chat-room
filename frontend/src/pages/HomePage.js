import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';
import { api } from '../services/api';

function HomePage() {
    const [sessions, setSessions] = useState([]);
    const [userSessions, setUserSessions] = useState([]);
    const [newSessionName, setNewSessionName] = useState('');
    const [user, setUser] = useState(authService.getUser());
    const navigate = useNavigate();

    useEffect(() => {
        fetchSessions();
        fetchUserSessions();
        const fetchUserData = async () => {
            const userData = await authService.fetchUserData();
            if (userData) {
                setUser(userData);
            }
        };
        fetchUserData();
    }, []);

    const fetchUserSessions = async () => {
        try {
            const data = await api.users.getSessions();
            if (Array.isArray(data)) {
                setUserSessions(data);
            }
        } catch (error) {
            console.error('Error fetching user sessions:', error);
        }
    };

    const fetchSessions = async () => {
        try {
            const data = await api.sessions.list();
            if (Array.isArray(data)) {
                setSessions(data);
            }
        } catch (error) {
            console.error('Error fetching sessions:', error);
        }
    };

    const createSession = async () => {
        if (!newSessionName.trim()) return;

        try {
            const newSession = await api.sessions.create({ name: newSessionName });
            setNewSessionName('');
            await fetchSessions();
            await fetchUserSessions();
            
            if (newSession && newSession.id) {
                // Get session token for the new session
                const tokenResponse = await api.sessions.getToken();
                if (tokenResponse && tokenResponse.token) {
                    navigate(`/chat/${newSession.id}`);
                }
            }
        } catch (error) {
            console.error('Error creating session:', error);
        }
    };

    const handleJoinSession = async (token) => {
        try {
            await api.sessions.join(token);
            await fetchUserSessions();
            
            // Get session info from the token
            const shareInfo = await api.sessions.getShareInfo(token);
            if (shareInfo && shareInfo.session_id) {
                // Get session token
                const tokenResponse = await api.sessions.getToken();
                if (tokenResponse && tokenResponse.token) {
                    navigate(`/chat/${shareInfo.session_id}`);
                }
            }
        } catch (error) {
            console.error('Error joining session:', error);
        }
    };

    const getUserSessionRole = (sessionId) => {
        const userSession = userSessions.find(s => s.session_id === sessionId);
        return userSession ? userSession.role : null;
    };

    return (
        <div className="container mx-auto px-4 py-8">
            <div className="flex justify-between items-center mb-8">
                <h1 className="text-3xl font-bold">Chat Rooms</h1>
                <div className="flex space-x-4">
                    <button
                        onClick={() => navigate('/profile')}
                        className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                    >
                        {user?.avatar_url ? (
                            <img
                                src={user.avatar_url}
                                alt="Profile"
                                className="w-6 h-6 rounded-full object-cover"
                            />
                        ) : (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                        )}
                    </button>
                </div>
            </div>

            <div className="mb-8">
                <h2 className="text-xl font-semibold mb-4">Your Sessions</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {sessions.filter(session => userSessions.some(us => us.session_id === session.id)).map(session => (
                        <div key={session.id} className="bg-white rounded-lg shadow-md p-4">
                            <div className="flex justify-between items-center mb-2">
                                <h3 className="text-lg font-medium">{session.name}</h3>
                                {getUserSessionRole(session.id) === 'creator' && (
                                    <button
                                        onClick={() => navigate(`/sessions/${session.id}/manage`)}
                                        className="text-gray-600 hover:text-gray-800"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                            <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
                                        </svg>
                                    </button>
                                )}
                            </div>
                            <p className="text-sm text-gray-600 mb-4">
                                {session.users?.length || 0} members
                            </p>
                            <button
                                onClick={() => handleJoinSession(session.token)}
                                className="w-full px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                            >
                                Enter Chat
                            </button>
                        </div>
                    ))}
                </div>
            </div>

            <div className="mb-8">
                <h2 className="text-xl font-semibold mb-4">Create New Session</h2>
                <div className="flex gap-4">
                    <input
                        type="text"
                        value={newSessionName}
                        onChange={(e) => setNewSessionName(e.target.value)}
                        placeholder="Enter session name"
                        className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <button
                        onClick={createSession}
                        className="px-6 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
                    >
                        Create
                    </button>
                </div>
            </div>
        </div>
    );
}

export default HomePage; 