import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import RandomNicknameGenerator from '../components/register/RandomNicknameGenerator';
import { authService } from '../services/auth';
import { API_ENDPOINTS } from '../services/api';

const UserPage = () => {
    const navigate = useNavigate();
    const [user, setUser] = useState(authService.getUser());
    const [isEditingNickname, setIsEditingNickname] = useState(false);
    const [isEditingUsername, setIsEditingUsername] = useState(false);
    const [newNickname, setNewNickname] = useState('');
    const [newUsername, setNewUsername] = useState('');
    const [selectedFile, setSelectedFile] = useState(null);
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        const fetchUserData = async () => {
            const userData = await authService.fetchUserData();
            if (userData) {
                setUser(userData);
            }
        };
        fetchUserData();
    }, []);

    const handleUpdateNickname = async () => {
        if (!newNickname || newNickname === user.nickname) {
            setIsEditingNickname(false);
            return;
        }

        setIsLoading(true);
        setError('');

        try {
            const response = await fetch(API_ENDPOINTS.USERS.UPDATE_NICKNAME(user.id), {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${authService.getToken()}`
                },
                body: JSON.stringify({ nickname: newNickname })
            });

            const data = await response.json();

            if (response.ok) {
                const updatedUser = { ...user, nickname: newNickname };
                authService.updateStoredUser(updatedUser);
                setUser(updatedUser);
                setIsEditingNickname(false);
            } else {
                setError(data.message || 'Failed to update nickname');
            }
        } catch (error) {
            setError('Unable to connect to server');
        } finally {
            setIsLoading(false);
        }
    };

    const handleUpdateUsername = async () => {
        if (!newUsername || newUsername === user.username) {
            setIsEditingUsername(false);
            return;
        }

        setIsLoading(true);
        setError('');

        try {
            const response = await fetch(API_ENDPOINTS.USERS.UPDATE_USERNAME(user.id), {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${authService.getToken()}`
                },
                body: JSON.stringify({ username: newUsername })
            });

            const data = await response.json();

            if (response.ok) {
                const updatedUser = { ...user, username: newUsername };
                authService.updateStoredUser(updatedUser);
                setUser(updatedUser);
                setIsEditingUsername(false);
            } else {
                setError(data.message || 'Failed to update username');
            }
        } catch (error) {
            setError('Unable to connect to server');
        } finally {
            setIsLoading(false);
        }
    };

    const handleFileSelect = (event) => {
        const file = event.target.files[0];
        if (file) {
            if (file.size > 5 * 1024 * 1024) { // 5MB limit
                setError('File size too large. Please choose a file under 5MB.');
                return;
            }
            setSelectedFile(file);
            setError('');
        }
    };

    const handleAvatarUpload = async () => {
        if (!selectedFile) return;

        setIsLoading(true);
        setError('');

        try {
            const formData = new FormData();
            formData.append('avatar', selectedFile);

            const response = await fetch(API_ENDPOINTS.AVATAR.UPLOAD, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                },
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                const updatedUser = { ...user, avatar_url: data.avatar_url };
                authService.updateStoredUser(updatedUser);
                setUser(updatedUser);
                setSelectedFile(null);
            } else {
                setError(data.message || 'Failed to upload avatar');
            }
        } catch (error) {
            setError('Unable to connect to server');
        } finally {
            setIsLoading(false);
        }
    };

    const handleLogout = () => {
        authService.logout();
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md mx-auto bg-white rounded-lg shadow-lg p-8">
                <div className="text-center mb-8">
                    <h2 className="text-3xl font-bold text-gray-900">Profile</h2>
                </div>

                <div className="space-y-6">
                    {/* Avatar Section */}
                    <div className="flex flex-col items-center space-y-4">
                        <div className="relative">
                            <div className="w-32 h-32 rounded-full overflow-hidden bg-gray-200">
                                {user?.avatar_url ? (
                                    <img
                                        src={user.avatar_url}
                                        alt="Profile"
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <div className="w-full h-full flex items-center justify-center text-gray-400">
                                        <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                        </svg>
                                    </div>
                                )}
                            </div>
                            <label className="absolute bottom-0 right-0 bg-blue-600 rounded-full p-2 cursor-pointer hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                                <input
                                    type="file"
                                    className="hidden"
                                    accept="image/*"
                                    onChange={handleFileSelect}
                                />
                            </label>
                        </div>
                        {selectedFile && (
                            <button
                                onClick={handleAvatarUpload}
                                disabled={isLoading}
                                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
                            >
                                {isLoading ? 'Uploading...' : 'Upload New Avatar'}
                            </button>
                        )}
                    </div>

                    {/* Username Section */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Username</label>
                        <div className="mt-1">
                            {isEditingUsername ? (
                                <div className="flex items-center space-x-2">
                                    <input
                                        type="text"
                                        value={newUsername}
                                        onChange={(e) => setNewUsername(e.target.value)}
                                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                                        placeholder="Enter new username"
                                    />
                                    <button
                                        onClick={handleUpdateUsername}
                                        disabled={isLoading}
                                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
                                    >
                                        {isLoading ? 'Saving...' : 'Save'}
                                    </button>
                                    <button
                                        onClick={() => setIsEditingUsername(false)}
                                        className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            ) : (
                                <div className="flex items-center justify-between">
                                    <p className="text-gray-900">{user?.username}</p>
                                    <button
                                        onClick={() => {
                                            setNewUsername(user?.username || '');
                                            setIsEditingUsername(true);
                                        }}
                                        className="text-blue-600 hover:text-blue-800"
                                    >
                                        Edit
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Nickname Section */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Nickname</label>
                        <div className="mt-1">
                            {isEditingNickname ? (
                                <div className="flex items-center space-x-2">
                                    <input
                                        type="text"
                                        value={newNickname}
                                        onChange={(e) => setNewNickname(e.target.value)}
                                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                                        placeholder="Enter new nickname"
                                    />
                                    <button
                                        onClick={handleUpdateNickname}
                                        disabled={isLoading}
                                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
                                    >
                                        {isLoading ? 'Saving...' : 'Save'}
                                    </button>
                                    <button
                                        onClick={() => setIsEditingNickname(false)}
                                        className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            ) : (
                                <div className="flex items-center justify-between">
                                    <p className="text-gray-900">{user?.nickname}</p>
                                    <button
                                        onClick={() => {
                                            setNewNickname(user?.nickname || '');
                                            setIsEditingNickname(true);
                                        }}
                                        className="text-blue-600 hover:text-blue-800"
                                    >
                                        Edit
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>

                    {error && (
                        <p className="mt-2 text-sm text-red-600">{error}</p>
                    )}

                    <div className="flex justify-between pt-6">
                        <button
                            onClick={() => navigate('/home')}
                            className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                        >
                            Back to Chat Rooms
                        </button>
                        <button
                            onClick={handleLogout}
                            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                        >
                            Logout
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default UserPage; 