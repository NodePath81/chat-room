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
- **Tailwind CSS** - Utility-first CSS framework
- **WebSocket API** - Real-time communication

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 12 or higher
- Docker/Podman

### Docker Setup

1. Build and run using Docker Compose:
```bash
docker-compose up --build
```

## API Endpoints

You can find API endpoints in the backend/main.go file.

### Environment Variables

#### Backend
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `JWT_SECRET` - Secret for JWT signing


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
