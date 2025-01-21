import React, { useState, useRef } from 'react';

function SendBar({ onSendMessage, onImageUpload }) {
    const [message, setMessage] = useState('');
    const fileInputRef = useRef(null);
    const [selectedImages, setSelectedImages] = useState([]);
    const [imagePreviewUrls, setImagePreviewUrls] = useState([]);

    const handleImageSelect = (event) => {
        const files = Array.from(event.target.files);
        if (files.length > 0) {
            setSelectedImages(prev => [...prev, ...files]);
            
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
        <div className="bg-white border-t shadow-sm">
            {imagePreviewUrls.length > 0 && (
                <div className="flex gap-2 p-3 overflow-x-auto border-b">
                    {imagePreviewUrls.map((url, index) => (
                        <div key={index} className="relative inline-block">
                            <img src={url} alt="Preview" className="h-20 w-20 object-cover rounded-lg" />
                            <button
                                onClick={() => {
                                    setSelectedImages(prev => prev.filter((_, i) => i !== index));
                                    setImagePreviewUrls(prev => prev.filter((_, i) => i !== index));
                                }}
                                className="absolute -top-2 -right-2 p-1 bg-white rounded-full shadow-md hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                    ))}
                </div>
            )}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex items-end space-x-3 py-3">
                    <button
                        onClick={() => fileInputRef.current?.click()}
                        className="p-2 rounded-full text-gray-600 hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                    </button>
                    <input
                        type="file"
                        ref={fileInputRef}
                        accept="image/*"
                        onChange={handleImageSelect}
                        className="hidden"
                        multiple
                    />
                    <div className="flex-1">
                        <textarea
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            onKeyDown={handleKeyPress}
                            placeholder="Type a message..."
                            className="w-full resize-none border rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-h-[44px] max-h-32"
                            rows="1"
                        />
                    </div>
                    <button
                        onClick={handleSendMessage}
                        disabled={!message.trim() && !selectedImages.length}
                        className={`p-2 rounded-full ${
                            message.trim() || selectedImages.length
                                ? 'bg-blue-600 hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
                                : 'bg-gray-300'
                        }`}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    );
}

export default SendBar;
