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
    const readyToFetchRef = useRef(false);

    const handleScroll = (e) => {
        const { scrollTop, scrollHeight, clientHeight } = e.target;
        const isContentShorterThanContainer = scrollHeight <= clientHeight;
        
        // Allow triggering update zone even with short content
        if (scrollTop < 50 || isContentShorterThanContainer) {
            if (!showUpdateZone) {
                onUpdateZoneChange(true, false);
                readyToFetchRef.current = false;
            } else if (!updateZoneExpanded && (scrollTop < lastScrollTopRef.current || isContentShorterThanContainer)) {
                onUpdateZoneChange(true, true);
                readyToFetchRef.current = true;
                
                // Reposition to show the oldest message
                if (messages.length > 0) {
                    const firstMessage = messageListRef.current.querySelector('[data-message-id]');
                    if (firstMessage) {
                        firstMessage.scrollIntoView({ block: 'start', behavior: 'smooth' });
                    }
                }
            }
        } else {
            onUpdateZoneChange(false, false);
            readyToFetchRef.current = false;
        }

        // Check if scrolled to top and ready to fetch
        if ((scrollTop === 0 || isContentShorterThanContainer) && hasMore && !isLoadingMore) {
            if (readyToFetchRef.current) {
                oldHeightRef.current = scrollHeight;
                onScroll();
                readyToFetchRef.current = false;
            }
        }
        
        // Update last scroll position
        lastScrollTopRef.current = scrollTop;
    };

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
            onScroll={handleScroll}
            className="flex-1 overflow-y-scroll h-full px-4 space-y-4"
            style={{ 
                scrollbarWidth: 'thin',
                scrollbarGutter: 'stable',
                scrollbarColor: '#CBD5E1 transparent',
                minHeight: '100%'
            }}
        >
            {isLoading ? (
                <div className="flex items-center justify-center h-full">
                    <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                </div>
            ) : (
                <>
                    <div 
                        ref={updateZoneRef}
                        className={`sticky top-0 left-0 right-0 transition-all duration-300 overflow-hidden bg-blue-50 rounded-lg ${
                            showUpdateZone ? 'mb-4' : 'mb-0'
                        } ${
                            updateZoneExpanded ? 'h-16 opacity-100' : 'h-0 opacity-0'
                        }`}
                        style={{
                            transform: updateZoneExpanded ? 'translateY(0)' : 'translateY(-100%)'
                        }}
                    >
                        <div className="flex items-center justify-center h-full text-blue-600">
                            {isLoadingMore ? (
                                <div className="flex items-center space-x-2">
                                    <div className="w-5 h-5 border-3 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                                    <span>Loading messages...</span>
                                </div>
                            ) : hasMore ? (
                                <span>{readyToFetchRef.current ? "Release to load more" : "Pull down to load more"}</span>
                            ) : (
                                <span className="text-gray-500">No more messages</span>
                            )}
                        </div>
                    </div>

                    {/* Ensure minimum height for short content */}
                    <div className="min-h-full">
                        {!hasMore && messages.length > 0 && (
                            <div className="text-center py-4 text-gray-500 text-sm">
                                Beginning of conversation
                            </div>
                        )}
                        
                        {messages.map((message, index) => (
                            <div key={message.id || index} data-message-id={message.id}>
                                <MessageBubble
                                    message={message}
                                    user={users[message.user_id]}
                                />
                            </div>
                        ))}
                        
                        <div ref={messagesEndRef} />
                    </div>
                </>
            )}
        </div>
    );
}

export default ChatBoard;
