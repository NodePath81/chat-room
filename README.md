# Real-time Chat Room

A real-time chat application built with Go (backend) and React (frontend), featuring WebSocket communication for instant messaging.

## Features

- 🚀 Real-time messaging using WebSocket
- 👥 Multiple chat rooms support
- 🔐 User authentication
- 👤 Online user presence
- 💾 Message persistence
- 🔄 Automatic reconnection
- 📱 Responsive design

## Tech Stack

### Backend
- Go 1.21+
- Chi (routing)
- GORM (database ORM)
- Gorilla WebSocket
- PostgreSQL

### Frontend
- React 18
- TypeScript
- Chakra UI
- Axios
- React Router DOM

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/chat-room
cd chat-room
```

2. **Set up the backend**
```bash
cd backend

# Install dependencies
go mod tidy

# Create .env file
cp .env.example .env

# Update database configuration in .env
# Start the server
go run main.go
```

3. **Set up the frontend**
```bash
cd frontend

# Install dependencies
npm install

# Start the development server
npm start
```

### Environment Variables

#### Backend (.env)
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=chatroom
JWT_SECRET=your_secret_key
```

#### Frontend (.env)
```env
REACT_APP_API_URL=http://localhost:8080
REACT_APP_WS_URL=ws://localhost:8080
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user

### Sessions
- `GET /api/sessions` - List all sessions
- `POST /api/sessions` - Create new session
- `GET /api/sessions/{id}` - Get session details
- `POST /api/sessions/{id}/join` - Join a session
- `DELETE /api/sessions/{id}/leave` - Leave a session

### WebSocket
- `WS /ws/session/{sessionId}` - WebSocket endpoint for chat sessions

## Project Structure

```
chat-app/
├── backend/
│   ├── main.go
│   ├── handlers/
│   │   ├── websocket.go
│   │   ├── session.go
│   │   └── auth.go
│   ├── models/
│   │   ├── user.go
│   │   ├── session.go
│   │   └── message.go
│   ├── database/
│   │   └── db.go
│   └── config/
│       └── config.go
└── frontend/
    ├── src/
    │   ├── components/
    │   │   ├── ChatRoom.tsx
    │   │   ├── MessageList.tsx
    │   │   ├── MessageInput.tsx
    │   │   └── UserList.tsx
    │   ├── services/
    │   │   └── websocket.ts
    │   └── App.tsx
    └── package.json
```

## Usage

1. Register a new account or login
2. Create a new chat room or join an existing one
3. Start chatting in real-time with other users

## WebSocket Message Format

```typescript
interface Message {
    type: 'message' | 'user_joined' | 'user_left';
    sessionId: string;
    content?: string;
    user?: {
        id: number;
        username: string;
    };
    timestamp: string;
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

- [Chi](https://github.com/go-chi/chi)
- [GORM](https://gorm.io)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [React](https://reactjs.org)
- [Chakra UI](https://chakra-ui.com)
```

This README.md provides:
- Clear installation instructions
- Project structure
- API endpoints
- Environment setup
- Tech stack details
- Usage instructions
- Contributing guidelines

