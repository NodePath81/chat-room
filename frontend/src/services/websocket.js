export class WebSocketService {
  constructor() {
    this.ws = null;
    this.handlers = new Map();
    this.reconnectTimeout = null;
    this.currentSessionId = null;
    this.isConnecting = false;
    this.maxRetries = 3;
    this.retryCount = 0;
  }

  connect(sessionId) {
    if (this.isConnecting) return;
    this.isConnecting = true;
    this.currentSessionId = parseInt(sessionId, 10);

    const token = localStorage.getItem('token');
    if (!token) {
        this.isConnecting = false;
        return;
    }

    const wsUrl = `ws://localhost:8080/ws?sessionId=${this.currentSessionId}`;
    
    try {
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('Connected to WebSocket for session:', this.currentSessionId);
            this.isConnecting = false;
            this.retryCount = 0;
            this.ws.send(JSON.stringify({
                type: 'auth',
                token: token
            }));
        };

        this.ws.onclose = (event) => {
            this.isConnecting = false;
            console.log('WebSocket connection closed:', event.code);
            
            if (this.retryCount < this.maxRetries) {
                this.retryCount++;
                const delay = Math.min(1000 * Math.pow(2, this.retryCount), 10000);
                this.reconnectTimeout = setTimeout(() => {
                    if (this.currentSessionId === parseInt(sessionId, 10)) {
                        this.connect(sessionId);
                    }
                }, delay);
            } else {
                console.log('Max retries reached, stopping reconnection');
            }
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'history') {
                    if (this.handlers.has('history')) {
                        this.handlers.get('history')(message.messages);
                    }
                    return;
                }
                
                if (message.type === 'auth_success') {
                    console.log('Authenticated successfully');
                    return;
                }

                if (message.error) {
                    console.error('WebSocket error:', message.error);
                    this.disconnect();
                    return;
                }

                if (this.handlers.has('message')) {
                    this.handlers.get('message')(message);
                }
            } catch (e) {
                console.error('Error parsing message:', e);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.isConnecting = false;
        };
    } catch (error) {
        console.error('Failed to connect:', error);
        this.isConnecting = false;
    }
  }

  sendMessage(content, sessionId) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        console.error('WebSocket not connected');
        return;
    }

    const parsedSessionId = parseInt(sessionId, 10);
    if (parsedSessionId !== this.currentSessionId) {
        console.error('Cannot send message to different session', parsedSessionId, this.currentSessionId);
        return;
    }

    const message = {
      type: 'message',
      content: content,
      sessionId: parsedSessionId
    };

    this.ws.send(JSON.stringify(message));
  }

  onMessage(callback) {
    this.handlers.set('message', callback);
  }

  disconnect() {
    this.isConnecting = false;
    this.retryCount = this.maxRetries; // Stop reconnection attempts
    
    if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = null;
    }

    this.handlers.clear();

    if (this.ws) {
        this.ws.onclose = null;
        this.ws.close();
        this.ws = null;
    }

    this.currentSessionId = null;
  }

  onHistory(callback) {
    this.handlers.set('history', callback);
  }
}

export default new WebSocketService(); 