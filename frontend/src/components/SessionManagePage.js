import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { API_ENDPOINTS, api } from '../services/api';
import { authService } from '../services/auth';

const SessionManagePage = () => {
    const { id: sessionId } = useParams();
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
        fetchSessionData();
    }, [sessionId]);

    const fetchSessionData = async () => {
        try {
            setLoading(true);
            setError('');

            // Check role first
            const roleData = await api.sessions.checkRole(sessionId);
            setRole(roleData.role);

            // If not creator, redirect to home
            if (roleData.role !== 'creator') {
                navigate('/');
                return;
            }

            // Fetch session details
            const sessionData = await api.sessions.get(sessionId);
            setSession(sessionData);

            // Fetch members
            const membersData = await api.sessions.listMembers(sessionId);
            setMembers(membersData.members);

        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateShareLink = async () => {
        try {
            setError('');
            const data = await api.sessions.createShareLink(sessionId, { durationDays: duration });
            const fullShareLink = `${window.location.origin}/share?token=${data.token}`;
            setShareLink(fullShareLink);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleKickMember = async (memberId) => {
        try {
            setError('');
            await api.sessions.kickMember(sessionId, memberId);
            // Refresh members list
            fetchSessionData();
        } catch (err) {
            setError(err.message);
        }
    };

    const handleRemoveSession = async () => {
        if (!window.confirm('Are you sure you want to remove this session? This action cannot be undone.')) {
            return;
        }

        try {
            setError('');
            await api.sessions.remove(sessionId);
            navigate('/');
        } catch (err) {
            setError(err.message);
        }
    };

    const fetchMemberDetails = async (memberId) => {
        try {
            const userData = await api.users.get(memberId);
            setMemberDetails(prev => ({
                ...prev,
                [memberId]: userData
            }));
        } catch (error) {
            console.error('Error fetching member details:', error);
        }
    };

    useEffect(() => {
        members.forEach(memberId => {
            if (!memberDetails[memberId]) {
                fetchMemberDetails(memberId);
            }
        });
    }, [members]);

    if (loading) {
        return <div className="flex justify-center items-center min-h-screen">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        </div>;
    }

    if (role !== 'creator') {
        return null; // Will redirect in useEffect
    }

    return (
        <div className="container mx-auto px-4 py-8">
            <div className="max-w-4xl mx-auto space-y-8">
                {/* Session Info */}
                <div className="bg-white shadow rounded-lg p-6">
                    <h1 className="text-2xl font-bold mb-4">Session Management</h1>
                    {session && (
                        <div className="mb-4">
                            <h2 className="text-xl font-semibold">{session.name}</h2>
                        </div>
                    )}
                </div>

                {/* Share Link Generator */}
                <div className="bg-white shadow rounded-lg p-6">
                    <h2 className="text-xl font-semibold mb-4">Create Share Link</h2>
                    <div className="space-y-4">
                        <div className="flex items-center space-x-4">
                            <label className="text-gray-700">Duration (days):</label>
                            <select 
                                value={duration} 
                                onChange={(e) => setDuration(Number(e.target.value))}
                                className="border rounded px-3 py-2"
                            >
                                {[1, 3, 7, 14, 30].map(days => (
                                    <option key={days} value={days}>{days} days</option>
                                ))}
                            </select>
                            <button
                                onClick={handleCreateShareLink}
                                className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
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
                                        className="flex-1 border rounded px-3 py-2 bg-gray-50"
                                    />
                                    <button
                                        onClick={() => {
                                            navigator.clipboard.writeText(shareLink);
                                            alert('Link copied to clipboard!');
                                        }}
                                        className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                                    >
                                        Copy
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Members List */}
                <div className="bg-white shadow rounded-lg p-6">
                    <h2 className="text-xl font-semibold mb-4">Members</h2>
                    <div className="space-y-4">
                        {members.map(memberId => {
                            const member = memberDetails[memberId] || {};
                            return (
                                <div key={memberId} className="flex items-center justify-between border-b py-3">
                                    <div className="flex items-center space-x-3">
                                        {member.avatarUrl ? (
                                            <img 
                                                src={member.avatarUrl} 
                                                alt={member.nickname} 
                                                className="w-10 h-10 rounded-full object-cover"
                                            />
                                        ) : (
                                            <div className="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center">
                                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                                </svg>
                                            </div>
                                        )}
                                        <span className="font-medium text-gray-900">
                                            {member.nickname || 'Loading...'}
                                        </span>
                                    </div>
                                    {memberId !== session?.creatorId && (
                                        <button
                                            onClick={() => handleKickMember(memberId)}
                                            className="text-red-500 hover:text-red-700 focus:outline-none"
                                        >
                                            Kick
                                        </button>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Danger Zone */}
                <div className="bg-red-50 shadow rounded-lg p-6">
                    <h2 className="text-xl font-semibold mb-4 text-red-700">Danger Zone</h2>
                    <button
                        onClick={handleRemoveSession}
                        className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                    >
                        Remove Session
                    </button>
                </div>

                {/* Error Display */}
                {error && (
                    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                        {error}
                    </div>
                )}
            </div>
        </div>
    );
};

export default SessionManagePage; 