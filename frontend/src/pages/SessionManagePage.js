import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import sessionService from '../services/session';
import { userService } from '../services/user';

const SessionManagePage = () => {
    const { sessionId } = useParams();
    const navigate = useNavigate();
    const [session, setSession] = useState(null);
    const [members, setMembers] = useState([]);
    const [memberDetails, setMemberDetails] = useState({});
    const [role, setRole] = useState('none');
    const [shareLink, setShareLink] = useState('');
    const [duration, setDuration] = useState(7);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        if (!sessionId) {
            navigate('/');
            return;
        }
        fetchSessionData();
    }, [sessionId, navigate]);

    const fetchSessionData = async () => {
        try {
            setLoading(true);
            setError('');

            // Initialize session and get details
            const sessionData = await sessionService.initializeSession(sessionId);
            setSession(sessionData.session);
            setRole(sessionData.role);

            // Fetch members
            const memberIds = await sessionService.getSessionMembers(sessionId);
            setMembers(memberIds);

            // Fetch member details
            const users = await userService.getBatchUsers(memberIds);
            const userDetailsMap = {};
            users.forEach(user => {
                if (user && user.id) {
                    userDetailsMap[user.id] = user;
                }
            });
            setMemberDetails(userDetailsMap);

        } catch (err) {
            console.error('Error fetching session data:', err);
            setError(err.message || 'Failed to load session data');
            if (err.message.includes('not found') || err.message.includes('verify')) {
                setTimeout(() => navigate('/'), 2000);
            }
        } finally {
            setLoading(false);
        }
    };

    const handleCreateShareLink = async () => {
        try {
            setError('');
            const data = await sessionService.createShareLink(sessionId, duration);
            if (!data || !data.token) {
                throw new Error('Failed to generate share link');
            }
            const fullShareLink = `${window.location.origin}/share?token=${data.token}`;
            setShareLink(fullShareLink);
        } catch (err) {
            console.error('Error creating share link:', err);
            setError(err.message || 'Failed to create share link');
        }
    };

    const handleKickMember = async (memberId) => {
        try {
            setError('');
            await sessionService.kickMember(sessionId, memberId);
            // Refresh members list
            const memberIds = await sessionService.getSessionMembers(sessionId);
            setMembers(memberIds);
            
            // Update member details
            const users = await userService.getBatchUsers(memberIds);
            const userDetailsMap = {};
            users.forEach(user => {
                if (user && user.id) {
                    userDetailsMap[user.id] = user;
                }
            });
            setMemberDetails(userDetailsMap);
        } catch (err) {
            console.error('Error kicking member:', err);
            setError(err.message || 'Failed to kick member');
        }
    };

    const handleRemoveSession = async () => {
        if (!window.confirm('Are you sure you want to remove this session? This action cannot be undone.')) {
            return;
        }

        try {
            setError('');
            await sessionService.removeSession(sessionId);
            navigate('/');
        } catch (err) {
            console.error('Error removing session:', err);
            setError(err.message || 'Failed to remove session');
        }
    };

    const handleLeaveSession = async () => {
        if (!window.confirm('Are you sure you want to leave this session?')) {
            return;
        }

        try {
            setError('');
            await sessionService.leaveSession(sessionId);
            navigate('/');
        } catch (err) {
            console.error('Error leaving session:', err);
            setError(err.message || 'Failed to leave session');
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-screen bg-gray-50">
                <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex justify-center items-center min-h-screen bg-gray-50">
                <div className="bg-white shadow-lg rounded-lg p-6 max-w-md w-full mx-4">
                    <div className="text-red-600 text-center mb-4">{error}</div>
                    <button
                        onClick={() => navigate('/')}
                        className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                    >
                        Return to Home
                    </button>
                </div>
            </div>
        );
    }

    if (!session) {
        return null;
    }

    return (
        <div className="min-h-screen bg-gray-50 py-8">
            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8">
                {/* Session Info */}
                <div className="bg-white shadow-sm rounded-lg p-6">
                    <div className="flex items-center justify-between mb-6">
                        <h1 className="text-2xl font-bold text-gray-900">Session Management</h1>
                        <button
                            onClick={() => navigate(`/sessions/${sessionId}`)}
                            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                            Back to Chat
                        </button>
                    </div>
                    {session && (
                        <div className="mb-4">
                            <h2 className="text-xl font-semibold text-gray-900">{session.name}</h2>
                        </div>
                    )}
                </div>

                {/* Share Link Generator - Only for creator */}
                {role === 'creator' && (
                    <div className="bg-white shadow-sm rounded-lg p-6">
                        <h2 className="text-xl font-semibold text-gray-900 mb-4">Create Share Link</h2>
                        <div className="space-y-4">
                            <div className="flex items-center space-x-4">
                                <label className="text-gray-700">Duration (days):</label>
                                <select 
                                    value={duration} 
                                    onChange={(e) => setDuration(Number(e.target.value))}
                                    className="border rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    {[1, 3, 7, 14, 30].map(days => (
                                        <option key={days} value={days}>{days} days</option>
                                    ))}
                                </select>
                                <button
                                    onClick={handleCreateShareLink}
                                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                                >
                                    Generate Link
                                </button>
                            </div>
                            {shareLink && (
                                <div className="mt-4">
                                    <label className="block text-gray-700 mb-2">Share Link:</label>
                                    <div className="flex items-center space-x-2">
                                        <input
                                            type="text"
                                            value={shareLink}
                                            readOnly
                                            className="flex-1 border rounded-md px-3 py-2 bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                        />
                                        <button
                                            onClick={() => {
                                                navigator.clipboard.writeText(shareLink);
                                                alert('Link copied to clipboard!');
                                            }}
                                            className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                                        >
                                            Copy
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                )}

                {/* Members List */}
                <div className="bg-white shadow-sm rounded-lg p-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">Members</h2>
                    <div className="space-y-4">
                        {members.map(memberId => {
                            const member = memberDetails[memberId] || {};
                            const isCreator = memberId === session?.creator_id;
                            return (
                                <div key={memberId} className="flex items-center justify-between py-3 border-b border-gray-200 last:border-0">
                                    <div className="flex items-center space-x-3">
                                        {member.avatar_url ? (
                                            <img 
                                                src={member.avatar_url} 
                                                alt={member.nickname} 
                                                className="w-10 h-10 rounded-full object-cover bg-gray-100"
                                            />
                                        ) : (
                                            <div className="w-10 h-10 rounded-full bg-blue-500 flex items-center justify-center text-white font-medium">
                                                {member.nickname ? member.nickname.charAt(0).toUpperCase() : '?'}
                                            </div>
                                        )}
                                        <div>
                                            <span className="font-medium text-gray-900">
                                                {member.nickname || 'Loading...'}
                                            </span>
                                            {isCreator && (
                                                <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                    Creator
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                    {role === 'creator' && !isCreator && (
                                        <button
                                            onClick={() => handleKickMember(memberId)}
                                            className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm leading-4 font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                                        >
                                            Kick
                                        </button>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Action Zone - Different for creator and member */}
                <div className={`${role === 'creator' ? 'bg-red-50' : 'bg-gray-50'} shadow-sm rounded-lg p-6`}>
                    <h2 className={`text-xl font-semibold ${role === 'creator' ? 'text-red-700' : 'text-gray-700'} mb-4`}>
                        {role === 'creator' ? 'Danger Zone' : 'Session Actions'}
                    </h2>
                    {role === 'creator' ? (
                        <button
                            onClick={handleRemoveSession}
                            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                        >
                            Remove Session
                        </button>
                    ) : (
                        <button
                            onClick={handleLeaveSession}
                            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-gray-600 hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500"
                        >
                            Leave Session
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};

export default SessionManagePage; 