export class WebSocketService {
  constructor(url) {
    this.url = url;
    this.ws = null;
  }

  connect(sessionId, token) {
    this.ws = new WebSocket(`${this.url}/ws/session/${sessionId}?token=${token}`);

    this.ws.onopen = () => {
      console.log('Connected to WebSocket');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log('Received message:', message);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from WebSocket');
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  sendMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'message',
        content: message,
      }));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
} 