export class WebSocketService {
  constructor() {
    this.ws = null;
    this.handlers = new Map();
  }

  connect(sessionId) {
    const token = localStorage.getItem('token');
    if (!token) return;

    const wsUrl = `ws://localhost:8080/ws?sessionId=${sessionId}`;
    
    try {
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('Connected to WebSocket');
            this.ws.send(JSON.stringify({
                type: 'auth',
                token: token
            }));
        };

        this.ws.onclose = () => {
            console.log('WebSocket connection closed');
            setTimeout(() => this.connect(sessionId), 2000);
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
        };
    } catch (error) {
        console.error('Failed to connect:', error);
    }
  }

  sendMessage(content, sessionId) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

    const message = {
      type: 'message',
      content: content,
      sessionId: sessionId
    };

    this.ws.send(JSON.stringify(message));
  }

  onMessage(callback) {
    this.handlers.set('message', callback);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  onHistory(callback) {
    this.handlers.set('history', callback);
  }
}

export default new WebSocketService(); 