import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';
import sessionService from '../services/session';

function HomePage() {
    const [sessions, setSessions] = useState([]);
    const [newSessionName, setNewSessionName] = useState('');
    const [user, setUser] = useState(authService.getUser());
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        fetchUserData();
        fetchSessions();
    }, []);

    const fetchUserData = async () => {
        try {
            const userData = await authService.fetchUserData();
            if (userData) {
                setUser(userData);
            }
        } catch (error) {
            console.error('Error fetching user data:', error);
            setError('Failed to load user data');
        }
    };

    const fetchSessions = async () => {
        try {
            setLoading(true);
            setError(null);
            const sessionIds = await sessionService.getSessionIDs();
            const sessionsData = await sessionService.getSessions(sessionIds);
            setSessions(sessionsData);
        } catch (error) {
            console.error('Error fetching sessions:', error);
            setError('Failed to load sessions');
        } finally {
            setLoading(false);
        }
    };

    const createSession = async () => {
        if (!newSessionName.trim()) return;

        try {
            setError(null);
            const newSession = await sessionService.createSession(newSessionName);
            setNewSessionName('');
            await fetchSessions();
            
            if (newSession && newSession.id) {
                navigate(`/sessions/${newSession.id}`);
            }
        } catch (error) {
            console.error('Error creating session:', error);
            setError('Failed to create session');
        }
    };

    const handleEnterChat = (sessionId) => {
        navigate(`/sessions/${sessionId}`);
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

            {error && (
                <div className="mb-4 p-4 bg-red-100 text-red-700 rounded-md">
                    {error}
                </div>
            )}

            <div className="mb-8">
                <h2 className="text-xl font-semibold mb-4">Your Sessions</h2>
                {loading ? (
                    <div className="flex justify-center items-center h-32">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                    </div>
                ) : sessions.length === 0 ? (
                    <div className="text-center text-gray-500 py-8">
                        No sessions found. Create a new one below!
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {sessions.map(session => (
                            <div key={session.id} className="bg-white rounded-lg shadow-md p-4">
                                <div className="flex justify-between items-center mb-2">
                                    <h3 className="text-lg font-medium">{session.name}</h3>
                                </div>
                                <p className="text-sm text-gray-600 mb-4">
                                    {session.users?.length || 0} members
                                    {session.userRole === 'creator' && (
                                        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                            Creator
                                        </span>
                                    )}
                                </p>
                                <button
                                    onClick={() => handleEnterChat(session.id)}
                                    className="w-full px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                >
                                    Enter Chat
                                </button>
                            </div>
                        ))}
                    </div>
                )}
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