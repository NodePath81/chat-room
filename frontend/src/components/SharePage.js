import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { API_ENDPOINTS } from '../services/api';
import { authService } from '../services/auth';

const SharePage = () => {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const joinSession = async () => {
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

        joinSession();
    }, [navigate, searchParams]);

    if (isLoading) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50">
                <div className="w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
                <div className="text-xl text-gray-600 font-semibold">Joining session...</div>
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

    return null;
};

export default SharePage; 