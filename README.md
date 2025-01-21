# Real-time Chat Application

A modern real-time chat application built with Go (backend) and React (frontend), featuring secure authentication, multiple chat rooms, real-time messaging, and image sharing capabilities.

## Features

- 🔐 **Secure Authentication**
  - JWT-based authentication
  - User registration and login
  - Protected routes and API endpoints

- 💬 **Real-time Chat**
  - WebSocket-based real-time messaging
  - Multiple chat rooms support
  - Message history with infinite scroll
  - Image sharing support
  - User avatars and nicknames
  - Auto-reconnection on connection loss
  - Message type support (text/image)
  - Real-time user presence

- 🎨 **Modern UI**
  - Clean and responsive design with Tailwind CSS
  - Intuitive message interface
  - Pull-to-load more messages
  - Image preview and upload
  - Loading states and animations
  - Error handling and user feedback

## Tech Stack

### Backend
- **Go** - Programming language
- **Chi** - Lightweight router
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

3. Set up environment variables:
```bash
cp .env.example .env
```

4. Start the development server:
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
- `GET /api/sessions/:id` - Get session details
- `POST /api/sessions/:id/join` - Join a chat session
- `GET /api/sessions/:id/messages` - Get session messages
- `POST /api/sessions/:id/messages` - Send a message
- `POST /api/sessions/:id/upload` - Upload an image

### Users
- `GET /api/users/:id` - Get user information
- `PUT /api/users/:id` - Update user profile
- `POST /api/users/:id/avatar` - Upload user avatar

### WebSocket
- `WS /ws` - WebSocket endpoint for real-time chat
  - Query Parameters:
    - `session_id`: ID of the chat session to join
    - `token`: JWT authentication token

## Project Structure

```
.
├── backend/
│   ├── config/         # Configuration
│   ├── database/       # Database setup
│   ├── handlers/       # HTTP & WebSocket handlers
│   ├── middleware/     # Authentication middleware
│   ├── models/         # Database models
│   ├── store/          # Data access layer
│   └── main.go         # Entry point
├── frontend/
│   ├── src/
│   │   ├── components/ # React components
│   │   │   ├── chat/   # Chat-related components
│   │   │   └── user/   # User-related components
│   │   ├── services/   # API & WebSocket services
│   │   └── App.js      # Root component
│   └── package.json
└── docker-compose.yml
```

### Environment Variables

#### Backend
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `JWT_SECRET` - Secret for JWT signing
- `UPLOAD_DIR` - Directory for file uploads

#### Frontend
- `REACT_APP_API_URL` - Backend API URL
- `REACT_APP_WS_URL` - WebSocket URL
- `REACT_APP_MAX_IMAGE_SIZE` - Maximum image upload size

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
