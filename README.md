
# README.md
# Restaurant Ordering System

A modern digital ordering system for restaurants with QR code table scanning, real-time order tracking, and integrated payment processing.

## Features

### Customer Features
- QR code scanning for table identification
- Digital menu browsing with categories
- 360° product view for menu items
- Shopping cart management
- Real-time order tracking
- Integrated payment via Midtrans
- Call waiter assistance

### Staff Features
- Order management dashboard
- Real-time notifications
- Menu management with media upload
- Payment verification
- Sales analytics
- Multi-role support (Admin, Cashier, Waiter, Kitchen)

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gorilla Mux
- **Database**: PostgreSQL
- **Cache**: Redis
- **Payment**: Midtrans
- **Real-time**: WebSocket
- **Authentication**: JWT

### Frontend
- **Customer App**: React.js PWA
- **Staff Dashboard**: React.js with TypeScript
- **State Management**: Redux Toolkit
- **UI**: Material-UI / Ant Design
- **Real-time**: Socket.io-client

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Node.js 18+
- Docker & Docker Compose (optional)

## Installation

### 1. Clone the repository
```bash
git clone https://github.com/your-repo/lendral3n/ordering-system.git
cd lendral3n/ordering-system
```

### 2. Backend Setup

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Install dependencies
go mod download

# Run database migrations
make migrate-up

# Seed initial data
make seed

# Run the server
make run
```

### 3. Frontend Setup

#### Customer App
```bash
cd frontend/customer-app
npm install
npm start
```

#### Staff Dashboard
```bash
cd frontend/staff-dashboard
npm install
npm start
```

### 4. Using Docker

```bash
# Build and run all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## API Documentation

API documentation is available at `/api/docs` when running in development mode.

## Project Structure

```
lendral3n/ordering-system-system/
├── cmd/api/              # Application entry point
├── internal/             # Private application code
│   ├── config/          # Configuration
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── routes/          # Route definitions
│   ├── services/        # Business logic
│   └── utils/           # Utilities
├── pkg/                 # Public packages
├── scripts/             # Utility scripts
├── frontend/            # Frontend applications
│   ├── customer-app/    # Customer PWA
│   └── staff-dashboard/ # Staff dashboard
└── deployments/         # Deployment configurations
```

## Environment Variables

Key environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT tokens
- `MIDTRANS_SERVER_KEY`: Midtrans server key
- `MIDTRANS_CLIENT_KEY`: Midtrans client key
- `REDIS_URL`: Redis connection string

See `.env.example` for complete list.

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/handlers/...
```

## Deployment

### Production Deployment

1. Set production environment variables
2. Build the application: `make build`
3. Run migrations: `make migrate-up`
4. Start the application with process manager (systemd, supervisor)

### Docker Deployment

```bash
# Build production image
docker build -t lendral3n/ordering-system:latest .

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

## Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/new-feature`
5. Submit pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, email support@lendral3n/ordering-system.com or create an issue in the repository.