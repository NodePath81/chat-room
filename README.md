# Real-time Chat Application

A modern real-time chat application built with Go (backend) and React (frontend), featuring secure authentication, multiple chat rooms, and real-time messaging capabilities.

## Features

- 🔐 **Secure Authentication**
  - JWT-based authentication
  - User registration and login
  - Protected routes and API endpoints

- 💬 **Real-time Chat**
  - WebSocket-based real-time messaging
  - Multiple chat rooms support
  - Message history persistence
  - Username display for messages
  - Auto-reconnection on connection loss

- 🎨 **Modern UI**
  - Clean and responsive design with Tailwind CSS
  - Intuitive user interface
  - Loading states and animations
  - Error handling and user feedback

## Tech Stack

### Backend
- **Go** - Programming language
- **Chi** - Lightweight router
- **GORM** - ORM for PostgreSQL
- **JWT-Go** - JWT authentication
- **Gorilla WebSocket** - WebSocket implementation
- **PostgreSQL** - Database

### Frontend
- **React** - UI library
- **React Router** - Client-side routing
- **Tailwind CSS** - Utility-first CSS framework
- **WebSocket API** - Real-time communication

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 12 or higher
- Docker (optional)

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=your_db_user
export DB_PASSWORD=your_db_password
export DB_NAME=chat_db
export JWT_SECRET=your_jwt_secret
```

4. Run the server:
```bash
go run main.go
```

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
yarn install
```

3. Start the development server:
```bash
yarn start
```

### Docker Setup

1. Build and run using Docker Compose:
```bash
docker-compose up --build
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and receive JWT token

### Sessions
- `GET /api/sessions` - Get all chat sessions
- `POST /api/sessions` - Create a new chat session
- `POST /api/sessions/:id/join` - Join a chat session
- `GET /api/sessions/:id/check` - Check session membership

### Users
- `GET /api/users/:id` - Get user information (protected)

### WebSocket
- `WS /ws` - WebSocket endpoint for real-time chat
  - Query Parameters:
    - `sessionId`: ID of the chat session to join

## Project Structure

```
.
├── backend/
│   ├── config/         # Configuration
│   ├── handlers/       # HTTP & WebSocket handlers
│   ├── middleware/     # Authentication middleware
│   ├── models/         # Database models
│   └── main.go         # Entry point
├── frontend/
│   ├── src/
│   │   ├── components/ # React components
│   │   ├── services/   # API & WebSocket services
│   │   └── App.js      # Root component
│   └── package.json
└── docker-compose.yml
```

## Development

### Running Tests
```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
yarn test
```

### Environment Variables

#### Backend
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `JWT_SECRET` - Secret for JWT signing

#### Frontend
- `REACT_APP_API_URL` - Backend API URL
- `REACT_APP_WS_URL` - WebSocket URL

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.