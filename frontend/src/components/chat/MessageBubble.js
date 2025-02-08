import React from 'react';

function MessageContent({ message }) {
    switch (message.type) {
        case 'image':
            return (
                <div className="mt-2">
                    <img
                        src={message.content}
                        alt="Message attachment"
                        className="max-w-sm rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer"
                        onClick={() => window.open(message.content, '_blank')}
                        onError={(e) => {
                            e.target.onerror = null;
                            e.target.src = '/default-image-error.png';
                            e.target.className = "w-16 h-16 opacity-50";
                        }}
                    />
                </div>
            );
        case 'text':
        default:
            return (
                <div className="mt-1 text-gray-800 break-words whitespace-pre-wrap">
                    {message.content}
                </div>
            );
    }
}

function MessageBubble({ message, user }) {
    const timestamp = new Date(message.timestamp).toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit'
    });

    const getInitial = (nickname) => {
        return nickname ? nickname.charAt(0).toUpperCase() : '?';
    };

    // Ensure we have the correct user data structure
    const userData = user || { nickname: 'Unknown User', avatar_url: null };

    return (
        <div className="flex items-start space-x-2 mb-4">
            <div className="flex-shrink-0 w-8 h-8">
                {userData.avatar_url ? (
                    <img
                        src={userData.avatar_url}
                        alt={userData.nickname}
                        className="w-8 h-8 rounded-full object-cover"
                    />
                ) : (
                    <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-medium">
                        {getInitial(userData.nickname)}
                    </div>
                )}
            </div>
            <div className="flex-1">
                <div className="flex items-baseline space-x-2">
                    <span className="font-medium text-gray-900">
                        {userData.nickname}
                    </span>
                    <span className="text-xs text-gray-500">{timestamp}</span>
                </div>
                <MessageContent message={message} />
            </div>
        </div>
    );
}

export default MessageBubble;
