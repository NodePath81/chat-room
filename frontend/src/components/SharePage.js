import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { API_ENDPOINTS } from '../services/api';
import { authService } from '../services/auth';

const SharePage = () => {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(true);
    const [sessionInfo, setSessionInfo] = useState(null);

    useEffect(() => {
        const fetchSessionInfo = async () => {
            const token = searchParams.get('token');
            if (!token) {
                setError('Invalid share link');
                setIsLoading(false);
                return;
            }

            if (!authService.isAuthenticated()) {
                navigate('/login', { 
                    state: { returnTo: `/share?${searchParams.toString()}` }
                });
                return;
            }

            try {
                const response = await fetch(`${API_ENDPOINTS.SESSIONS.GET_SHARE_INFO}?token=${token}`, {
                    headers: {
                        'Authorization': `Bearer ${authService.getToken()}`
                    }
                });

                if (!response.ok) {
                    throw new Error('Failed to get session information');
                }

                const data = await response.json();
                setSessionInfo(data);
                setIsLoading(false);
            } catch (error) {
                console.error('Error fetching session info:', error);
                setError('Failed to get session information. Please try again.');
                setIsLoading(false);
            }
        };

        fetchSessionInfo();
    }, [navigate, searchParams]);

    const handleJoin = async () => {
        const token = searchParams.get('token');
        setIsLoading(true);

        try {
            const response = await fetch(`${API_ENDPOINTS.SESSIONS.JOIN}?token=${token}`, {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error('Failed to join session');
            }

            const data = await response.json();
            navigate(`/chat/${data.session_id}`);
        } catch (error) {
            console.error('Error joining session:', error);
            setError('Failed to join session. Please try again.');
            setIsLoading(false);
        }
    };

    const handleCancel = () => {
        navigate('/');
    };

    if (isLoading) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50">
                <div className="w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
                <div className="text-xl text-gray-600 font-semibold">Loading session information...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50">
                <div className="bg-white p-8 rounded-lg shadow-md max-w-md w-full">
                    <div className="text-red-500 text-center mb-4">{error}</div>
                    <button
                        onClick={() => navigate('/')}
                        className="w-full px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                    >
                        Go to Home
                    </button>
                </div>
            </div>
        );
    }

    if (sessionInfo) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="bg-white p-8 rounded-lg shadow-md max-w-md w-full">
                    <div className="text-center mb-8">
                        <h2 className="text-2xl font-bold text-gray-900 mb-4">Session Invitation</h2>
                        <p className="text-lg text-gray-700">
                            <span className="font-semibold text-blue-600">{sessionInfo.inviter_nickname}</span>
                            {' '}invites you to join{' '}
                            <span className="font-semibold text-blue-600">{sessionInfo.session_name}</span>
                        </p>
                    </div>
                    
                    <div className="space-y-4">
                        <div className="flex justify-between space-x-4">
                            <button
                                onClick={handleCancel}
                                className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleJoin}
                                className="flex-1 px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                            >
                                Join Session
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return null;
};

export default SharePage; 