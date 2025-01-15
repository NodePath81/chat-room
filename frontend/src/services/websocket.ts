export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  connect(sessionId: string, token: string) {
    this.ws = new WebSocket(`${this.url}/ws/session/${sessionId}?token=${token}`);

    this.ws.onopen = () => {
      console.log('Connected to WebSocket');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      // Handle incoming messages
      console.log('Received message:', message);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from WebSocket');
      // Implement reconnection logic here
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  sendMessage(message: string) {
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