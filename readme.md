# go-auth-api

A production-style REST API built in Go featuring JWT authentication, 
refresh token rotation, bcrypt password hashing, rate limiting, and 
structured logging.

## Tech Stack

- **Language:** Go 1.22+
- **Database:** PostgreSQL
- **Auth:** JWT (golang-jwt) + bcrypt
- **Email:** Brevo API
- **Migrations:** golang-migrate

## Features

- User registration with bcrypt password hashing
- JWT authentication with refresh token rotation
- Custom HTTP middleware — logger, auth, rate limiter
- Structured JSON logging with slog
- Background goroutine for expired token cleanup
- Posts resource with per-user ownership enforcement
- Environment-based configuration

## Getting Started

### Prerequisites
- Go 1.22+
- PostgreSQL
- golang-migrate CLI

### Setup

```bash
# Clone the repo
git clone https://github.com/thushAR13/go-auth-api
cd go-auth-api

# Install dependencies
go mod download

# Create .env file
cp .env.example .env
# Fill in your values

# Run migrations
migrate -path migrations -database "postgres://..." up

# Start the server
go run .
```

## API Endpoints

### Auth
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/register | No | Register a new user |
| POST | /api/login | No | Login, returns tokens |
| POST | /api/refresh | No | Refresh access token |
| POST | /api/logout | No | Invalidate refresh token |

### Posts
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/posts | Yes | Create a post |
| GET | /api/posts | Yes | Get your posts |
| GET | /api/posts/{id} | Yes | Get one post |
| DELETE | /api/posts/{id} | Yes | Delete your post |

## Example Usage

```bash
# Register
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Create a post
curl -X POST http://localhost:8080/api/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello","body":"My first post"}'
```