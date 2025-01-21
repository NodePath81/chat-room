import { API_ENDPOINTS } from './api';

export class WebSocketService {
  constructor() {
    this.connections = new Map(); // sessionId -> { ws, messageHandlers, reconnectAttempts, shouldReconnect }
    this.maxReconnectAttempts = 5;
  }

  connect(sessionId) {
    // If connection already exists for this session, close it
    if (this.connections.has(sessionId)) {
      this.disconnect(sessionId);
    }

    const connection = {
      ws: null,
      messageHandlers: [],
      reconnectAttempts: 0,
      shouldReconnect: true
    };

    this.connections.set(sessionId, connection);

    const setupWebSocket = () => {
      connection.ws = new WebSocket(API_ENDPOINTS.WEBSOCKET.CONNECT(sessionId));

      connection.ws.onopen = () => {
        // Send authentication message
        const token = localStorage.getItem('token');
        connection.ws.send(JSON.stringify({ token }));
        connection.reconnectAttempts = 0;
        console.log('WebSocket connection established for session:', sessionId);
      };

      connection.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('Received WebSocket message:', data);
          connection.messageHandlers.forEach(handler => handler(data));
        } catch (error) {
          console.error('Error handling WebSocket message:', error);
        }
      };

      connection.ws.onclose = (event) => {
        console.log('WebSocket connection closed:', event.code, event.reason);
        // Only attempt to reconnect if shouldReconnect is true and connection still exists
        if (connection.shouldReconnect && 
            this.connections.has(sessionId) && 
            connection.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            connection.reconnectAttempts++;
            console.log('Attempting to reconnect WebSocket, attempt:', connection.reconnectAttempts);
            setupWebSocket();
          }, Math.min(1000 * Math.pow(2, connection.reconnectAttempts), 30000));
        } else {
          // Clean up the connection if we're not reconnecting
          this.connections.delete(sessionId);
        }
      };

      connection.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        if (connection.ws) {
          connection.ws.close();
        }
      };
    };

    setupWebSocket();
  }

  disconnect(sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      // Set shouldReconnect to false before closing
      connection.shouldReconnect = false;
      if (connection.ws) {
        connection.ws.close();
      }
      this.connections.delete(sessionId);
    }
  }

  disconnectAll() {
    for (const [sessionId] of this.connections) {
      this.disconnect(sessionId);
    }
  }

  sendMessage(message) {
    console.log('Sending message:', message);
    const connection = this.connections.get(message.sessionId);
    if (connection && connection.ws && connection.ws.readyState === WebSocket.OPEN) {
      console.log('Connection found and ready, sending message');
      connection.ws.send(JSON.stringify(message));
    } else {
      console.error('Failed to send message:', {
        hasConnection: !!connection,
        hasWs: connection?.ws ? 'yes' : 'no',
        readyState: connection?.ws?.readyState,
        expectedState: WebSocket.OPEN
      });
    }
  }

  onMessage(handler, sessionId) {
    console.log('Registering message handler for session:', sessionId);
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers.push(handler);
    } else {
      console.error('No connection found for session:', sessionId);
    }
  }

  removeHandlers(sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers = [];
    }
  }
}

export default new WebSocketService(); 