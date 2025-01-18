import { API_ENDPOINTS } from '../config';

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
      connection.ws = new WebSocket(`${process.env.REACT_APP_WS_URL || 'ws://localhost:8080'}/ws?sessionId=${sessionId}`);

      connection.ws.onopen = () => {
        // Send authentication message
        const token = localStorage.getItem('token');
        connection.ws.send(JSON.stringify({ token }));
        connection.reconnectAttempts = 0;
      };

      connection.ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        connection.messageHandlers.forEach(handler => handler(data));
      };

      connection.ws.onclose = () => {
        // Only attempt to reconnect if shouldReconnect is true and connection still exists
        if (connection.shouldReconnect && 
            this.connections.has(sessionId) && 
            connection.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            connection.reconnectAttempts++;
            setupWebSocket();
          }, Math.min(1000 * Math.pow(2, connection.reconnectAttempts), 30000));
        } else {
          // Clean up the connection if we're not reconnecting
          this.connections.delete(sessionId);
        }
      };

      connection.ws.onerror = () => {
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