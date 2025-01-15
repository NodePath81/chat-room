import React, { useState, useEffect } from 'react';
import { 
  Box, 
  VStack, 
  Heading, 
  Button, 
  Input, 
  useDisclosure, 
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  List,
  ListItem,
  Text
} from '@chakra-ui/react';
import { useNavigate } from 'react-router-dom';

function HomePage() {
  const [sessions, setSessions] = useState([]);
  const [newSessionName, setNewSessionName] = useState('');
  const { isOpen, onOpen, onClose } = useDisclosure();
  const navigate = useNavigate();

  useEffect(() => {
    fetchSessions();
  }, []);

  const fetchSessions = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/sessions');
      const data = await response.json();
      setSessions(data);
    } catch (error) {
      console.error('Error fetching sessions:', error);
    }
  };

  const createSession = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/sessions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: newSessionName }),
      });
      const data = await response.json();
      onClose();
      setNewSessionName('');
      fetchSessions();
    } catch (error) {
      console.error('Error creating session:', error);
    }
  };

  return (
    <Box p={8}>
      <VStack spacing={8} align="stretch">
        <Heading>Chat Rooms</Heading>
        
        <Button colorScheme="blue" onClick={onOpen}>
          Create New Chat Room
        </Button>

        <List spacing={3}>
          {sessions.map((session) => (
            <ListItem 
              key={session.ID}
              p={4}
              bg="white"
              borderRadius="md"
              shadow="sm"
              _hover={{ bg: 'gray.50', cursor: 'pointer' }}
              onClick={() => navigate(`/chat/${session.ID}`)}
            >
              <Text fontSize="lg">{session.name}</Text>
              <Text fontSize="sm" color="gray.500">
                Created at: {new Date(session.CreatedAt).toLocaleString()}
              </Text>
            </ListItem>
          ))}
        </List>

        <Modal isOpen={isOpen} onClose={onClose}>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Create New Chat Room</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <Input
                placeholder="Enter room name"
                value={newSessionName}
                onChange={(e) => setNewSessionName(e.target.value)}
              />
            </ModalBody>
            <ModalFooter>
              <Button colorScheme="blue" mr={3} onClick={createSession}>
                Create
              </Button>
              <Button variant="ghost" onClick={onClose}>Cancel</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </VStack>
    </Box>
  );
}

export default HomePage; 