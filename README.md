# Realtime Project Chat

A proof-of-concept project for learning event-driven architecture with Kafka, WebSockets, and real-time collaboration features

## 🎯 Learning Focus

This is a **learning project** focused on:

- **Kafka pub/sub messaging** for event-driven architecture
- **WebSocket real-time communication** for live chat
- **Go backend**
- **React TypeScript frontend**
- **Event sourcing** and real-time data synchronization

## ✨ Features

- **User Authentication** - JWT-based auth with bcrypt password hashing
- **Project Management** - Create projects and invite team members
- **Task Management** - Kanban-style task organization
- **Real-time Chat** - WebSocket-powered instant messaging between project members
- **Event-driven Architecture** - Kafka integration for pub/sub messaging

## 🚀 Quick Start

### Prerequisites

- Go 1.25
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL
- Kafka

### Setup

1. **Start infrastructure**

   ```bash
   docker-compose up -d
   ```

2. **Backend setup**

   ```bash
   cd backend
   cp .env.example .env
   make goose-up
   go run ./cmd/api
   ```

3. **Frontend setup**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:3333

## 🛠️ Technology Stack

### Backend

- **Go** with Chi router and clean architecture
- **PostgreSQL** with pgx driver and SQLC for type-safe queries
- **Kafka** with Sarama client for pub/sub messaging
- **JWT** authentication with secure token handling
- **WebSockets** for real-time communication
- **Docker** for development environment

### Frontend

- **React 19** with TypeScript
- **TanStack Router** for type-safe routing
- **TanStack Query** for server state management
- **Tailwind CSS** for styling
- **Radix UI** for accessible components
- **Vite** for fast development builds

## 🧪 Development

This project serves as a learning playground for:

- Event-driven communication
- Real-time data synchronization patterns
- WebSocket connection management
- Kafka producer/consumer patterns
- Modern React patterns with TypeScript

**Note**: This is a proof-of-concept project for educational purposes. Not intended for production use.
