import React, { useRef, useEffect } from 'react';
import MessageBubble from './MessageBubble';

function ChatBoard({
    messages,
    users,
    isLoading,
    isLoadingMore,
    hasMore,
    showUpdateZone,
    updateZoneExpanded,
    onScroll,
    onUpdateZoneChange,
}) {
    const messageListRef = useRef(null);
    const updateZoneRef = useRef(null);
    const messagesEndRef = useRef(null);
    const lastScrollTopRef = useRef(0);
    const oldHeightRef = useRef(0);

    useEffect(() => {
        const messageList = messageListRef.current;
        if (messageList) {
            const handleScroll = () => {
                const { scrollTop, scrollHeight, clientHeight } = messageList;
                
                // Update zone visibility based on scroll position
                if (scrollTop < 20) {
                    if (!showUpdateZone) {
                        onUpdateZoneChange(true, false);
                    } else if (!updateZoneExpanded && scrollTop < lastScrollTopRef.current) {
                        onUpdateZoneChange(true, true);
                    }
                } else if (scrollTop > 50) {
                    onUpdateZoneChange(false, false);
                }

                // Check if scrolled to top for loading more messages
                if (scrollTop === 0 && hasMore && !isLoadingMore) {
                    oldHeightRef.current = scrollHeight;
                    onScroll();
                }
                
                // Update last scroll position
                lastScrollTopRef.current = scrollTop;
            };

            messageList.addEventListener('scroll', handleScroll);
            return () => messageList.removeEventListener('scroll', handleScroll);
        }
    }, [hasMore, isLoadingMore, onScroll, showUpdateZone, updateZoneExpanded, onUpdateZoneChange]);

    // Preserve scroll position when new messages are loaded at the top
    useEffect(() => {
        const messageList = messageListRef.current;
        if (messageList && oldHeightRef.current) {
            const newHeight = messageList.scrollHeight;
            const heightDiff = newHeight - oldHeightRef.current;
            if (heightDiff > 0) {
                messageList.scrollTop = heightDiff;
            }
            oldHeightRef.current = 0;
        }
    }, [messages]);

    // Scroll to bottom for new messages
    useEffect(() => {
        const messageList = messageListRef.current;
        if (messageList && messages.length > 0) {
            const isAtBottom = messageList.scrollHeight - messageList.clientHeight <= messageList.scrollTop + 100;
            if (isAtBottom) {
                messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
            }
        }
    }, [messages]);

    return (
        <div
            ref={messageListRef}
            className="flex-1 overflow-y-auto p-4 space-y-4"
        >
            {isLoading ? (
                <div className="flex items-center justify-center h-full">
                    <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                </div>
            ) : (
                <>
                    <div 
                        ref={updateZoneRef}
                        className={`sticky top-0 left-0 right-0 transition-all duration-300 overflow-hidden ${
                            showUpdateZone ? 'mb-4' : 'mb-0'
                        } ${
                            updateZoneExpanded ? 'h-16 opacity-100' : 'h-0 opacity-0'
                        }`}
                        style={{
                            transform: updateZoneExpanded ? 'translateY(0)' : 'translateY(-100%)'
                        }}
                    >
                        <div className="flex items-center justify-center h-full bg-blue-50 rounded-lg">
                            {isLoadingMore ? (
                                <div className="flex items-center space-x-2">
                                    <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                                    <span className="text-blue-600">Loading more messages...</span>
                                </div>
                            ) : hasMore ? (
                                <span className="text-blue-600">
                                    {updateZoneExpanded ? 'Loading more messages...' : 'Scroll up to load more'}
                                </span>
                            ) : (
                                <span className="text-gray-500">No more messages</span>
                            )}
                        </div>
                    </div>

                    <div className="space-y-4 min-h-full">
                        {messages.map((msg, index) => (
                            msg && msg.content && (
                                <div key={msg.id || index} data-message-id={msg.id}>
                                    <MessageBubble 
                                        message={msg} 
                                        user={users[msg.user_id]} 
                                    />
                                </div>
                            )
                        ))}
                    </div>
                    <div ref={messagesEndRef} />
                </>
            )}
        </div>
    );
}

export default ChatBoard;
