import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';

function UserPage() {
    const [user, setUser] = useState(authService.getUser());
    const [isEditing, setIsEditing] = useState(false);
    const [selectedFile, setSelectedFile] = useState(null);
    const [previewUrl, setPreviewUrl] = useState(null);
    const [isUploading, setIsUploading] = useState(false);
    const [error, setError] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        // Fetch latest user data when component mounts
        const fetchUserData = async () => {
            try {
                const userData = await authService.fetchUserData();
                if (userData) {
                    setUser(userData);
                    // If user has an avatar URL, use it
                    if (userData.avatarUrl) {
                        setPreviewUrl(userData.avatarUrl);
                    }
                }
            } catch (err) {
                setError('Failed to fetch user data');
                console.error('Error fetching user data:', err);
            }
        };
        fetchUserData();
    }, []);

    const handleFileSelect = (event) => {
        const file = event.target.files[0];
        if (file) {
            if (file.size > 5 * 1024 * 1024) { // 5MB limit
                setError('File size too large. Please choose a file under 5MB.');
                return;
            }
            setSelectedFile(file);
            setPreviewUrl(URL.createObjectURL(file));
            setError(null);
        }
    };

    const handleAvatarUpload = async () => {
        if (!selectedFile) return;

        setIsUploading(true);
        setError(null);
        try {
            const formData = new FormData();
            formData.append('avatar', selectedFile);

            const response = await fetch('http://localhost:8080/api/avatar', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`
                },
                body: formData
            });

            if (!response.ok) {
                throw new Error('Failed to upload avatar');
            }

            const data = await response.json();
            
            // Update both local state and stored user data
            const updatedUser = { ...user, avatarUrl: data.avatarUrl };
            setUser(updatedUser);
            setPreviewUrl(data.avatarUrl);
            authService.updateStoredUser(updatedUser);
            
            // Clean up the file input
            setSelectedFile(null);
        } catch (error) {
            setError('Failed to upload avatar. Please try again.');
            console.error('Error uploading avatar:', error);
        } finally {
            setIsUploading(false);
            setIsEditing(false);
        }
    };

    const handleLogout = () => {
        authService.logout();
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-3xl mx-auto">
                <div className="bg-white shadow-lg rounded-lg overflow-hidden">
                    {/* Header */}
                    <div className="px-6 py-4 bg-blue-600">
                        <div className="flex justify-between items-center">
                            <h1 className="text-2xl font-bold text-white">User Profile</h1>
                            <button
                                onClick={handleLogout}
                                className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                            >
                                Logout
                            </button>
                        </div>
                    </div>

                    {/* Profile Content */}
                    <div className="p-6">
                        <div className="flex flex-col items-center space-y-4">
                            {/* Error Message */}
                            {error && (
                                <div className="w-full p-3 bg-red-100 text-red-700 rounded-md">
                                    {error}
                                </div>
                            )}

                            {/* Avatar */}
                            <div className="relative group">
                                <div className="w-32 h-32 rounded-full overflow-hidden bg-gray-200">
                                    {previewUrl ? (
                                        <img
                                            src={previewUrl}
                                            alt="Profile"
                                            className="w-full h-full object-cover"
                                            onError={(e) => {
                                                console.error('Error loading image:', e);
                                                setError('Failed to load avatar image');
                                                setPreviewUrl(null);
                                            }}
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center text-gray-400">
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                            </svg>
                                        </div>
                                    )}
                                </div>
                                {isEditing && (
                                    <div className="absolute inset-0 flex items-center justify-center">
                                        <label className="cursor-pointer bg-black bg-opacity-50 rounded-full p-2 text-white hover:bg-opacity-70">
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
                                )}
                            </div>

                            {/* Debug Info - Remove in production */}
                            <div className="text-xs text-gray-500">
                                <p>Avatar URL: {user?.avatarUrl || 'No avatar URL'}</p>
                                <p>Preview URL: {previewUrl || 'No preview URL'}</p>
                            </div>

                            {/* User Info */}
                            <div className="text-center">
                                <h2 className="text-xl font-semibold text-gray-800">{user?.username}</h2>
                                <p className="text-gray-500">User ID: {user?.id}</p>
                            </div>

                            {/* Actions */}
                            <div className="flex space-x-4">
                                {isEditing ? (
                                    <>
                                        <button
                                            onClick={handleAvatarUpload}
                                            disabled={!selectedFile || isUploading}
                                            className={`px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 ${
                                                (!selectedFile || isUploading) ? 'opacity-50 cursor-not-allowed' : ''
                                            }`}
                                        >
                                            {isUploading ? 'Uploading...' : 'Save Changes'}
                                        </button>
                                        <button
                                            onClick={() => {
                                                setIsEditing(false);
                                                setPreviewUrl(user?.avatarUrl || null);
                                                setSelectedFile(null);
                                                setError(null);
                                            }}
                                            className="px-4 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                                        >
                                            Cancel
                                        </button>
                                    </>
                                ) : (
                                    <button
                                        onClick={() => setIsEditing(true)}
                                        className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                    >
                                        Edit Profile
                                    </button>
                                )}
                            </div>

                            {/* Navigation */}
                            <div className="mt-6">
                                <button
                                    onClick={() => navigate('/')}
                                    className="text-blue-500 hover:text-blue-600 focus:outline-none"
                                >
                                    ← Back to Chat Rooms
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default UserPage; 