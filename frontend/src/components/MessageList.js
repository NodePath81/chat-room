import React from 'react';
import { VStack, Box, Text } from '@chakra-ui/react';

function MessageList({ messages }) {
  return (
    <VStack
      flex={1}
      w="full"
      overflowY="auto"
      spacing={4}
      align="stretch"
      p={4}
      bg="gray.50"
      borderRadius="md"
    >
      {messages.map((message, index) => (
        <Box
          key={index}
          bg="white"
          p={3}
          borderRadius="lg"
          shadow="sm"
        >
          <Text fontWeight="bold">{message.user?.username || 'Anonymous'}</Text>
          <Text>{message.content}</Text>
          <Text fontSize="xs" color="gray.500">
            {new Date(message.timestamp).toLocaleString()}
          </Text>
        </Box>
      ))}
    </VStack>
  );
}

export default MessageList; 