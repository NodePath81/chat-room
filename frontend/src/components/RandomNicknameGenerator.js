import React, { useState, useEffect } from 'react';
import { commonWords } from '../utils/words';

const RandomNicknameGenerator = ({ onNicknameChange, className, showValidation, validationStatus, isLoading }) => {
    const [placeholder, setPlaceholder] = useState('');
    const [hasUserTyped, setHasUserTyped] = useState(false);
    const [value, setValue] = useState('');

    const generateRandomNickname = () => {
        const word1 = commonWords[Math.floor(Math.random() * commonWords.length)];
        const word2 = commonWords[Math.floor(Math.random() * commonWords.length)];
        const numbers = Math.floor(Math.random() * 1000).toString().padStart(3, '0');
        return `${word1}${word2}${numbers}`;
    };

    useEffect(() => {
        const newNickname = generateRandomNickname();
        setPlaceholder(newNickname);
        if (!hasUserTyped) {
            onNicknameChange(newNickname);
        }
    }, []);

    const handleInputChange = (e) => {
        const newValue = e.target.value;
        setValue(newValue);
        setHasUserTyped(true);
        onNicknameChange(newValue);
    };

    const handleGenerateClick = () => {
        const newNickname = generateRandomNickname();
        if (hasUserTyped) {
            setValue(newNickname);
        } else {
            setPlaceholder(newNickname);
        }
        onNicknameChange(newNickname);
    };

    const ValidationIcon = () => {
        if (!showValidation) return null;
        if (isLoading) return <span className="text-gray-400 ml-2">⟳</span>;
        if (validationStatus === null) return null;
        return validationStatus ? 
            <span className="text-green-500 ml-2">✓</span> : 
            <span className="text-red-500 ml-2">✗</span>;
    };

    return (
        <div className="relative">
            <input
                type="text"
                placeholder={placeholder}
                value={value}
                onChange={handleInputChange}
                className={`w-full px-4 py-2 border rounded-md ${className}`}
            />
            <div className="absolute right-2 top-1/2 transform -translate-y-1/2 flex items-center">
                <ValidationIcon />
                <button
                    onClick={handleGenerateClick}
                    className="ml-2 text-gray-500 hover:text-gray-700"
                    type="button"
                >
                    🎲
                </button>
            </div>
        </div>
    );
};

export default RandomNicknameGenerator; 