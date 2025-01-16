# Real-time Chat Application

A modern real-time chat application built with Go (backend) and React (frontend), featuring WebSocket communication and JWT authentication.

## Features

- 🔐 Secure user authentication with JWT
- 💬 Real-time messaging using WebSocket
- 👥 Multiple chat rooms support
- 🔄 Message history persistence
- 🚀 Auto-reconnection on connection loss
- 📱 Responsive UI with Chakra UI

## Tech Stack

### Backend
- Go
- Chi (HTTP router)
- GORM (ORM)
- PostgreSQL (Database)
- Gorilla WebSocket
- JWT for authentication

### Frontend
- React
- Chakra UI
- React Router
- WebSocket API

## Getting Started

### Prerequisites
- Go 1.19 or higher
- Node.js 16 or higher
- PostgreSQL
- Yarn package manager

### Backend Setup
1. Clone the repository
```bash
git clone <repository-url>
cd wschat
```

2. Set up the database
```bash
# Create a PostgreSQL database named 'chat'
createdb chat
```

3. Configure environment variables
```bash
# Create .env file in backend directory
cp backend/.env.example backend/.env
# Edit the .env file with your database credentials
```

4. Run the backend
```bash
cd backend
go mod download
go run main.go
```

### Frontend Setup
1. Install dependencies
```bash
cd frontend
yarn install
```

2. Start the development server
```bash
yarn start
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token

### Sessions (Protected Routes)
- `GET /api/users/{id}` - Get public user information

- `GET /api/sessions` - Get all chat sessions
- `POST /api/sessions` - Create a new chat session
- `GET /api/sessions/{id}` - Get session details
- `POST /api/sessions/{id}/join` - Join a chat session
- `GET /api/sessions/{id}/check` - Check session membership

### WebSocket
- `WS /ws` - WebSocket endpoint for real-time chat
  - Requires authentication token
  - Requires session ID as query parameter

## Project Structure

### Backend
```
backend/
├── auth/         # Authentication related code
├── config/       # Configuration management
├── database/     # Database setup and migrations
├── handlers/     # HTTP and WebSocket handlers
├── middleware/   # Custom middlewares
├── models/       # Database models
└── main.go       # Application entry point
```

### Frontend
```
frontend/
├── public/
├── src/
│   ├── components/   # React components
│   ├── services/     # API and WebSocket services
│   ├── App.js        # Main application component
│   └── index.js      # Entry point
└── package.json
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
