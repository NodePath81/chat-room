import { API_ENDPOINTS } from './api';

class WebSocketService {
  constructor() {
    this.ws = null;
    this.messageCallback = null;
    this.errorCallback = null;
    this.connectCallback = null;
    this.disconnectCallback = null;
    
    // Bind methods to maintain context
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.sendMessage = this.sendMessage.bind(this);
    this.onMessage = this.onMessage.bind(this);
    this.onError = this.onError.bind(this);
    this.onConnect = this.onConnect.bind(this);
    this.onDisconnect = this.onDisconnect.bind(this);
  }

  connect(sessionId) {
    if (this.ws) {
      this.disconnect();
    }

    // Use the new WebSocket endpoint that's protected by middleware
    this.ws = new WebSocket(`ws://localhost:8080/api/sessions/wschat?session_id=${sessionId}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      if (this.connectCallback) {
        this.connectCallback();
      }
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (this.messageCallback) {
        this.messageCallback(message);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      if (this.errorCallback) {
        this.errorCallback(error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      if (this.disconnectCallback) {
        this.disconnectCallback();
      }
    };
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  sendMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  onMessage(callback) {
    this.messageCallback = callback;
  }

  onError(callback) {
    this.errorCallback = callback;
  }

  onConnect(callback) {
    this.connectCallback = callback;
  }

  onDisconnect(callback) {
    this.disconnectCallback = callback;
  }
}

// Export singleton instance
export const websocketService = new WebSocketService(); 