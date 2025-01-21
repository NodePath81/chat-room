import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import RandomNicknameGenerator from '../components/register/RandomNicknameGenerator';
import { API_ENDPOINTS } from '../services/api';

const RegisterPage = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        username: '',
        nickname: '',
        password: ''
    });
    const [validation, setValidation] = useState({
        username: null,  // null = not checked, true = valid, false = invalid
        nickname: null
    });
    const [loading, setLoading] = useState({
        username: false,
        nickname: false
    });

    // Debounce function
    const debounce = (func, wait) => {
        let timeout;
        return (...args) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func(...args), wait);
        };
    };

    // Check username availability
    const checkUsername = async (username) => {
        if (!username) {
            setValidation(prev => ({ ...prev, username: null }));
            return;
        }
        setLoading(prev => ({ ...prev, username: true }));
        try {
            const response = await fetch(API_ENDPOINTS.AUTH.CHECK_USERNAME + `?username=${encodeURIComponent(username)}`);
            setValidation(prev => ({ ...prev, username: response.status === 200 }));
        } catch (error) {
            console.error('Error checking username:', error);
            setValidation(prev => ({ ...prev, username: false }));
        } finally {
            setLoading(prev => ({ ...prev, username: false }));
        }
    };

    // Check nickname availability
    const checkNickname = async (nickname) => {
        if (!nickname) {
            setValidation(prev => ({ ...prev, nickname: null }));
            return;
        }
        setLoading(prev => ({ ...prev, nickname: true }));
        try {
            const response = await fetch(API_ENDPOINTS.AUTH.CHECK_NICKNAME + `?nickname=${encodeURIComponent(nickname)}`);
            setValidation(prev => ({ ...prev, nickname: response.status === 200 }));
        } catch (error) {
            console.error('Error checking nickname:', error);
            setValidation(prev => ({ ...prev, nickname: false }));
        } finally {
            setLoading(prev => ({ ...prev, nickname: false }));
        }
    };

    // Debounced check functions
    const debouncedCheckUsername = debounce(checkUsername, 500);
    const debouncedCheckNickname = debounce(checkNickname, 500);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!validation.username || !validation.nickname) {
            return;
        }

        try {
            const response = await fetch(API_ENDPOINTS.AUTH.REGISTER, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData),
            });

            if (response.ok) {
                navigate('/login');
            } else {
                const data = await response.json();
                alert(data.message || 'Registration failed');
            }
        } catch (error) {
            console.error('Error during registration:', error);
            alert('Registration failed');
        }
    };

    const ValidationIcon = ({ isValid, isLoading }) => {
        if (isLoading) return <span className="text-gray-400">⟳</span>;
        if (isValid === null) return null;
        return isValid ? 
            <span className="text-green-500">✓</span> : 
            <span className="text-red-500">✗</span>;
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow-lg">
                <div>
                    <h2 className="text-center text-3xl font-extrabold text-gray-900">
                        Register
                    </h2>
                </div>
                <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        {/* Username field */}
                        <div className="relative">
                            <input
                                type="text"
                                required
                                className="w-full px-4 py-2 border rounded-md"
                                placeholder="Username"
                                value={formData.username}
                                onChange={(e) => {
                                    setFormData(prev => ({ ...prev, username: e.target.value }));
                                    debouncedCheckUsername(e.target.value);
                                }}
                            />
                            <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                                <ValidationIcon 
                                    isValid={validation.username}
                                    isLoading={loading.username}
                                />
                            </div>
                        </div>

                        {/* Nickname field with generator */}
                        <div className="relative">
                            <RandomNicknameGenerator
                                onNicknameChange={(nickname) => {
                                    setFormData(prev => ({ ...prev, nickname }));
                                    debouncedCheckNickname(nickname);
                                }}
                                showValidation={true}
                                validationStatus={validation.nickname}
                                isLoading={loading.nickname}
                            />
                        </div>

                        {/* Password field */}
                        <div>
                            <input
                                type="password"
                                required
                                className="w-full px-4 py-2 border rounded-md"
                                placeholder="Password"
                                value={formData.password}
                                onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                            />
                        </div>
                    </div>

                    <div>
                        <button
                            type="submit"
                            className="w-full py-2 px-4 border border-transparent rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                            disabled={!validation.username || !validation.nickname}
                        >
                            Register
                        </button>
                    </div>
                </form>
                <div className="text-center">
                    <button
                        onClick={() => navigate('/login')}
                        className="text-blue-600 hover:text-blue-800"
                    >
                        Already have an account? Login
                    </button>
                </div>
            </div>
        </div>
    );
};

export default RegisterPage; 