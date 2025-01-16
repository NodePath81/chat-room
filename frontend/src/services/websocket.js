import { API_ENDPOINTS } from '../config';

export class WebSocketService {
  constructor() {
    this.connections = new Map(); // sessionId -> { ws, messageHandlers, historyHandlers }
    this.maxReconnectAttempts = 5;
  }

  connect(sessionId) {
    // If connection already exists for this session, close it
    if (this.connections.has(sessionId)) {
      this.disconnect(sessionId);
    }

    const connection = {
      ws: new WebSocket(`${process.env.REACT_APP_WS_URL || 'ws://localhost:8080'}/ws?sessionId=${sessionId}`),
      messageHandlers: [],
      historyHandlers: [],
      reconnectAttempts: 0
    };

    connection.ws.onopen = () => {
      // Send authentication message
      const token = localStorage.getItem('token');
      connection.ws.send(JSON.stringify({ token }));
      connection.reconnectAttempts = 0;
    };

    connection.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'history') {
        // Sort messages by timestamp
        const sortedMessages = data.messages.sort((a, b) => 
          new Date(a.timestamp) - new Date(b.timestamp)
        );
        connection.historyHandlers.forEach(handler => handler(sortedMessages));
      } else {
        connection.messageHandlers.forEach(handler => handler(data));
      }
    };

    connection.ws.onclose = () => {
      if (connection.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => {
          connection.reconnectAttempts++;
          this.connect(sessionId);
        }, Math.min(1000 * Math.pow(2, connection.reconnectAttempts), 30000));
      }
    };

    this.connections.set(sessionId, connection);
  }

  disconnect(sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.ws.close();
      this.connections.delete(sessionId);
    }
  }

  disconnectAll() {
    for (const [sessionId] of this.connections) {
      this.disconnect(sessionId);
    }
  }

  sendMessage(content, sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection && connection.ws.readyState === WebSocket.OPEN) {
      connection.ws.send(JSON.stringify({ content }));
    }
  }

  onMessage(handler, sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers.push(handler);
    }
  }

  onHistory(handler, sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.historyHandlers.push(handler);
    }
  }

  removeHandlers(sessionId) {
    const connection = this.connections.get(sessionId);
    if (connection) {
      connection.messageHandlers = [];
      connection.historyHandlers = [];
    }
  }
}

export default new WebSocketService(); 