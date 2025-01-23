import { API_ENDPOINTS, api } from './api';

class WebSocketService {
  constructor() {
    this.ws = null;
    this.messageCallback = null;
    this.errorCallback = null;
    this.connectCallback = null;
    this.disconnectCallback = null;
    this.sessionId = null;
    
    // Bind methods to maintain context
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.sendMessage = this.sendMessage.bind(this);
    this.onMessage = this.onMessage.bind(this);
    this.onError = this.onError.bind(this);
    this.onConnect = this.onConnect.bind(this);
    this.onDisconnect = this.onDisconnect.bind(this);
    this.getWebSocketToken = this.getWebSocketToken.bind(this);
  }

  async getWebSocketToken() {
    console.debug('Requesting WebSocket token...');
    try {
      // Ensure we have a valid session token first
      await api.sessions.getToken(this.sessionId);
      console.debug('Session token verified');

      // Request WebSocket token
      const response = await api.sessions.getWsToken(this.sessionId);
      console.debug('WebSocket token obtained');
      return response.token;
    } catch (error) {
      console.error('Failed to obtain WebSocket token:', error);
      throw error;
    }
  }

  async connect(sessionId) {
    console.debug(`Initiating WebSocket connection for session ${sessionId}...`);
    this.sessionId = sessionId;

    if (this.ws) {
      console.debug('Closing existing WebSocket connection');
      this.disconnect();
    }

    try {
      // Get WebSocket token
      const wsToken = await this.getWebSocketToken();
      console.debug('WebSocket token obtained, establishing connection...');

      // Create WebSocket connection with token
      this.ws = new WebSocket(API_ENDPOINTS.WEBSOCKET.CONNECT(sessionId, wsToken));

      this.ws.onopen = () => {
        console.debug('WebSocket connection established successfully');
        if (this.connectCallback) {
          this.connectCallback();
        }
      };

      this.ws.onmessage = (event) => {
        console.debug('WebSocket message received:', event.data);
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
        console.debug('WebSocket connection closed');
        if (this.disconnectCallback) {
          this.disconnectCallback();
        }
      };
    } catch (error) {
      console.error('Failed to establish WebSocket connection:', error);
      if (this.errorCallback) {
        this.errorCallback(error);
      }
    }
  }

  disconnect() {
    if (this.ws) {
      console.debug('Disconnecting WebSocket...');
      this.ws.close();
      this.ws = null;
      this.sessionId = null;
    }
  }

  sendMessage(message) {
    if (!this.ws) {
      const error = new Error('WebSocket instance not initialized');
      console.error(error);
      throw error;
    }

    if (this.ws.readyState !== WebSocket.OPEN) {
      const error = new Error(`WebSocket not in OPEN state. Current state: ${this.ws.readyState}`);
      console.error(error);
      throw error;
    }

    if (!this.sessionId) {
      const error = new Error('No session ID available');
      console.error(error);
      throw error;
    }

    try {
      
      console.debug('Preparing to send message:', message);
      const messageStr = JSON.stringify(message);
      console.debug('Serialized message:', messageStr);
      
      this.ws.send(messageStr);
      console.debug('Message sent successfully');
    } catch (error) {
      console.error('Failed to send message:', error);
      throw error;
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