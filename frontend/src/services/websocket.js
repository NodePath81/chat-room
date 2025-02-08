import { API_ENDPOINTS } from './api';
import sessionService from './session';

class WebSocketService {
  constructor() {
    this.ws = null;
    this.messageCallback = null;
    this.errorCallback = null;
    this.connectCallback = null;
    this.disconnectCallback = null;
    this.sessionId = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectTimeout = null;
    
    // Bind methods
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.sendMessage = this.sendMessage.bind(this);
    this.onMessage = this.onMessage.bind(this);
    this.onError = this.onError.bind(this);
    this.onConnect = this.onConnect.bind(this);
    this.onDisconnect = this.onDisconnect.bind(this);
    this.reconnect = this.reconnect.bind(this);
  }

  async connect(sessionId) {
    console.debug(`Initiating WebSocket connection for session ${sessionId}...`);
    this.sessionId = sessionId;
    this.reconnectAttempts = 0;

    if (this.ws) {
      console.debug('Closing existing WebSocket connection');
      this.disconnect();
    }

    try {
      // Get WebSocket token using session service
      const wsToken = await sessionService.getWebSocketToken(sessionId);
      console.debug('WebSocket token obtained, establishing connection...');

      // Create WebSocket connection with token
      this.ws = new WebSocket(API_ENDPOINTS.WEBSOCKET.CONNECT(wsToken));

      this.ws.onopen = () => {
        console.debug('WebSocket connection established successfully');
        this.reconnectAttempts = 0;
        if (this.connectCallback) {
          this.connectCallback();
        }
      };

      this.ws.onmessage = (event) => {
        console.debug('WebSocket message received:', event.data);
        try {
          const message = JSON.parse(event.data);
          if (this.messageCallback) {
            this.messageCallback(message);
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        if (this.errorCallback) {
          this.errorCallback(error);
        }
      };

      this.ws.onclose = (event) => {
        console.debug('WebSocket connection closed:', event);
        if (this.disconnectCallback) {
          this.disconnectCallback(event);
        }
        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000) {
          this.reconnect();
        }
      };
    } catch (error) {
      console.error('Failed to establish WebSocket connection:', error);
      if (this.errorCallback) {
        this.errorCallback(error);
      }
      // Attempt to reconnect on connection failure
      this.reconnect();
    }
  }

  async reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    
    console.debug(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    
    clearTimeout(this.reconnectTimeout);
    this.reconnectTimeout = setTimeout(() => {
      if (this.sessionId) {
        this.connect(this.sessionId);
      }
    }, delay);
  }

  disconnect() {
    clearTimeout(this.reconnectTimeout);
    if (this.ws) {
      console.debug('Disconnecting WebSocket...');
      this.ws.close(1000, 'Normal closure');
      this.ws = null;
      this.sessionId = null;
      this.reconnectAttempts = 0;
    }
  }

  sendMessage(message) {
    if (!this.ws) {
      throw new Error('WebSocket instance not initialized');
    }

    if (this.ws.readyState !== WebSocket.OPEN) {
      throw new Error(`WebSocket not in OPEN state. Current state: ${this.ws.readyState}`);
    }

    try {
      console.debug('Sending message:', message);
      this.ws.send(JSON.stringify(message));
      console.debug('Message sent successfully');
    } catch (error) {
      console.error('Failed to send message:', error);
      throw error;
    }
  }

  // Event handlers
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

export const websocketService = new WebSocketService(); 