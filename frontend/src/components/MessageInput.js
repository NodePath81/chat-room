import React, { useState } from 'react';
import { HStack, Input, IconButton } from '@chakra-ui/react';
import { ArrowForwardIcon } from '@chakra-ui/icons';

function MessageInput({ onSendMessage }) {
  const [message, setMessage] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (message.trim()) {
      onSendMessage(message);
      setMessage('');
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ width: '100%' }}>
      <HStack w="full">
        <Input
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Type a message..."
          size="lg"
        />
        <IconButton
          type="submit"
          aria-label="Send message"
          icon={<ArrowForwardIcon />}
          size="lg"
          colorScheme="blue"
        />
      </HStack>
    </form>
  );
}

export default MessageInput; 