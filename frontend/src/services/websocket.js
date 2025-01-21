import { API_ENDPOINTS } from './api';

class WebSocketService {
  constructor() {
    this.connections = new Map(); // sessionId -> { ws, messageHandlers }
    
    // Bind methods to maintain 'this' context
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.disconnectAll = this.disconnectAll.bind(this);
    this.sendMessage = this.sendMessage.bind(this);
    this.onMessage = this.onMessage.bind(this);
    this.removeHandlers = this.removeHandlers.bind(this);
  }

  connect(sessionId) {
    if (this.connections.has(sessionId)) {
      this.disconnect(sessionId);
    }

    try {
      const ws = new WebSocket(`ws://localhost:8080/ws?sessionId=${sessionId}`);
      const connection = {
        ws,
        messageHandlers: []
      };

      this.connections.set(sessionId, connection);

      ws.onopen = () => {
        const token = localStorage.getItem('token');
        ws.send(JSON.stringify({ token }));
        console.log('WebSocket connected for session:', sessionId);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('Received WebSocket message:', data);
          connection.messageHandlers.forEach(handler => handler(data));
        } catch (error) {
          console.error('Error handling WebSocket message:', error);
        }
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected for session:', sessionId);
        this.connections.delete(sessionId);
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.disconnect(sessionId);
      };
    } catch (error) {
      console.error('Error establishing WebSocket connection:', error);
      this.connections.delete(sessionId);
    }
  }

  disconnect(sessionId) {
    console.log('Disconnecting WebSocket for session:', sessionId);
    const connection = this.connections.get(sessionId);
    if (connection) {
      if (connection.ws) {
        connection.ws.close();
      }
      this.connections.delete(sessionId);
    }
  }

  disconnectAll() {
    console.log('Disconnecting all WebSocket connections');
    for (const [sessionId] of this.connections) {
      this.disconnect(sessionId);
    }
  }

  sendMessage(message) {
    const connection = this.connections.get(message.session_id);
    if (connection && connection.ws && connection.ws.readyState === WebSocket.OPEN) {
      console.log('Sending WebSocket message:', message);
      connection.ws.send(JSON.stringify(message));
    } else {
      console.error('Failed to send message: WebSocket not connected');
      throw new Error('WebSocket not connected');
    }
  }

  onMessage(handler, sessionId) {
    console.log('Registering message handler for session:', sessionId);
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers.push(handler);
    } else {
      console.error('No connection found for session:', sessionId);
      throw new Error('No WebSocket connection found');
    }
  }

  removeHandlers(sessionId) {
    console.log('Removing message handlers for session:', sessionId);
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers = [];
    }
  }
}

// Create a singleton instance
const webSocketService = new WebSocketService();

// Export the instance methods directly
export const {
  connect,
  disconnect,
  disconnectAll,
  sendMessage,
  onMessage,
  removeHandlers
} = webSocketService; 