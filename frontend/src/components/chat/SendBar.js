import React, { useState, useRef } from 'react';

function SendBar({ onSendMessage, onImageUpload }) {
    const [message, setMessage] = useState('');
    const fileInputRef = useRef(null);

    const [selectedImages, setSelectedImages] = useState([]);
    const [imagePreviewUrls, setImagePreviewUrls] = useState([]);

    const handleImageSelect = (event) => {
        const files = Array.from(event.target.files);
        if (files.length > 0) {
            // Add new images to existing ones
            setSelectedImages(prev => [...prev, ...files]);
            
            // Generate preview URLs for new images
            files.forEach(file => {
                const reader = new FileReader();
                reader.onloadend = () => {
                    setImagePreviewUrls(prev => [...prev, reader.result]);
                };
                reader.readAsDataURL(file);
            });
        }
    };

    const handleSendMessage = () => {
        if (message.trim()) {
            onSendMessage(message);
            setMessage('');
        }

        selectedImages.forEach(image => {
            onImageUpload(image);
        });
        
        setSelectedImages([]);
        setImagePreviewUrls([]);
        
        
    };

    const handleKeyPress = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSendMessage();
        }
    };

    return (
        <div className="border-t bg-white">
            {imagePreviewUrls.length > 0 && (
                <div className="flex gap-2 p-2 overflow-x-auto">
                    {imagePreviewUrls.map((url, index) => (
                        <div key={index} className="relative inline-block">
                            <img src={url} alt="Preview" className="max-h-32 rounded-lg" />
                            <button
                                onClick={() => {
                                    setSelectedImages(prev => prev.filter((_, i) => i !== index));
                                    setImagePreviewUrls(prev => prev.filter((_, i) => i !== index));
                                }}
                                className="absolute -top-2 -right-2 p-1 bg-gray-100 rounded-full hover:bg-gray-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                    ))}
                </div>
            )}
            <div className="flex items-center space-x-2 p-2">
                <input
                    type="file"
                    ref={fileInputRef}
                    accept="image/*"
                    onChange={handleImageSelect}
                    className="hidden"
                />
                <button
                    onClick={() => fileInputRef.current?.click()}
                    className="p-2 rounded-full hover:bg-gray-100"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                </button>
                <textarea
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    onKeyDown={handleKeyPress}
                    placeholder="Type a message..."
                    className="flex-1 resize-none border rounded-lg p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    rows="1"
                />
                <button
                    onClick={handleSendMessage}
                    disabled={!message.trim() && !selectedImages.length}
                    className={`p-2 rounded-full ${
                        message.trim() || selectedImages.length
                            ? 'bg-blue-500 hover:bg-blue-600'
                            : 'bg-gray-300'
                    }`}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                </button>
            </div>
        </div>
    );
}

export default SendBar;
