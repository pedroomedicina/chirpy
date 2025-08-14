# Chirpy 🐦

A modern, lightweight social media API built with Go that allows users to post short messages called "chirps" (similar to tweets). Chirpy features user authentication, JWT-based sessions, profanity filtering, and premium user upgrades through webhook integration.

## 🚀 Features

- **User Management**: Create accounts, login/logout, and update profiles
- **Chirp System**: Post, read, and delete short messages (max 140 characters)
- **Authentication**: JWT-based access tokens with refresh token support
- **Profanity Filter**: Automatic content moderation for inappropriate language
- **Premium Users**: Chirpy Red upgrade system via webhook integration
- **RESTful API**: Clean, well-structured HTTP endpoints
- **PostgreSQL Database**: Robust data persistence with SQLC for type-safe queries
- **Admin Panel**: Metrics tracking and development tools

## 🛠️ Tech Stack

- **Language**: Go 1.23.2
- **Database**: PostgreSQL
- **Authentication**: JWT tokens with bcrypt password hashing
- **Database Queries**: SQLC for type-safe SQL
- **Migrations**: Goose for database schema management
- **Environment**: dotenv for configuration

## 📋 Prerequisites

Before running Chirpy, ensure you have:

- Go 1.23.2 or later
- PostgreSQL database
- Git

## 🔧 Installation & Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/pedroomedicina/chirpy.git
   cd chirpy
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   Create a `.env` file in the root directory:
   ```env
   DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
   JWT_SECRET=your-super-secret-jwt-key
   PLATFORM=dev
   POLKA_KEY=your-polka-webhook-key
   ```

4. **Set up the database**
   ```bash
   # Install goose for migrations
   go install github.com/pressly/goose/v3/cmd/goose@latest
   
   # Run migrations
   goose -dir sql/schema postgres "your-db-url" up
   ```

5. **Generate database code**
   ```bash
   # Install sqlc
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   
   # Generate type-safe database code
   sqlc generate
   ```

6. **Build and run**
   ```bash
   go build -o chirpy
   ./chirpy
   ```

The server will start on `http://localhost:8080`

## 📚 API Documentation

### Base URL
```
http://localhost:8080
```

### Authentication
Most endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Health Check
```http
GET /api/healthz
```
Returns server status.

#### User Management

**Create User**
```http
POST /api/users
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Login**
```http
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

Response includes access token and refresh token:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "refresh_token_here",
  "is_chirpy_red": false
}
```

**Update User**
```http
PUT /api/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

**Refresh Token**
```http
POST /api/refresh
Authorization: Bearer <refresh-token>
```

**Revoke Token**
```http
POST /api/revoke
Authorization: Bearer <refresh-token>
```

#### Chirp Management

**Create Chirp**
```http
POST /api/chirps
Authorization: Bearer <token>
Content-Type: application/json

{
  "body": "This is my first chirp!"
}
```

**Get All Chirps**
```http
GET /api/chirps?sort=asc&author_id=<user-id>
```

Query parameters:
- `sort`: `asc` or `desc` (default: `asc`)
- `author_id`: Filter by specific user (optional)

**Get Chirp by ID**
```http
GET /api/chirps/{id}
```

**Delete Chirp**
```http
DELETE /api/chirps/{id}
Authorization: Bearer <token>
```

#### Admin Endpoints

**View Metrics**
```http
GET /admin/metrics
```

**Reset Database** (dev only)
```http
POST /admin/reset
```

#### Webhooks

**Polka Webhook** (Premium Upgrades)
```http
POST /api/polka/webhooks
Authorization: ApiKey <polka-key>
Content-Type: application/json

{
  "event": "user.upgraded",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE
);
```

### Chirps Table
```sql
CREATE TABLE chirps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
```

### Refresh Tokens Table
```sql
CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);
```

## 🔒 Security Features

- **Password Hashing**: Uses bcrypt for secure password storage
- **JWT Authentication**: Stateless authentication with configurable expiration
- **Refresh Tokens**: Secure token renewal mechanism with expiration and revocation
- **Input Validation**: Request validation and sanitization
- **Profanity Filter**: Automatic filtering of inappropriate content
- **Authorization**: Users can only delete their own chirps

## 🏗️ Project Structure

```
chirpy/
├── main.go                 # Application entry point
├── types.go               # Data structures
├── api_handlers.go        # Core API handlers
├── chirp_handlers.go      # Chirp-specific handlers
├── session_handlers.go    # Authentication handlers
├── internal/
│   ├── auth/             # Authentication utilities
│   └── database/         # Generated database code
├── sql/
│   ├── schema/           # Database migrations
│   └── queries/          # SQL queries
├── public/               # Static files
└── sqlc.yaml            # SQLC configuration
```

## 🚀 Deployment

### Environment Variables for Production
```env
DB_URL=postgres://user:pass@host:5432/chirpy?sslmode=require
JWT_SECRET=your-production-jwt-secret-key
PLATFORM=production
POLKA_KEY=your-production-polka-key
```

### Building for Production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chirpy .

# Run migrations in production
goose -dir sql/schema postgres "$DB_URL" up
```

## 🧪 Development

### Running Tests
```bash
go test ./...
```

### Database Migrations
```bash
# Create new migration
goose -dir sql/schema create migration_name sql

# Apply migrations
goose -dir sql/schema postgres "$DB_URL" up

# Rollback migration
goose -dir sql/schema postgres "$DB_URL" down
```

### Regenerating Database Code
After modifying SQL queries:
```bash
sqlc generate
```

## 📝 API Response Examples

### Successful Chirp Creation
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "body": "This is my first chirp!",
  "user_id": "456e7890-e89b-12d3-a456-426614174000"
}
```

### Error Response
```json
{
  "error": "Chirp too long"
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is part of the Boot.dev curriculum and is for educational purposes.

## 🙏 Acknowledgments

- Built as part of the [Boot.dev](https://boot.dev) Go course
- Uses the excellent [SQLC](https://sqlc.dev/) for type-safe database queries
- Database migrations powered by [Goose](https://github.com/pressly/goose)
