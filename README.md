# MableTask API

A Go-based analytics and authentication API server using Gin, PostgreSQL, and ClickHouse.

## Features

- **User Authentication**: Signup, login, and logout endpoints with JWT-based authentication.
- **Event Tracking**: Track custom events and analytics data, stored in ClickHouse for high performance.
- **Analytics Endpoints**: Retrieve event counts, average event duration, custom parameter analytics, unique users, and top page paths.
- **User Profile**: Secure endpoint to fetch user profile and IP address.
- **CORS Middleware**: Configurable CORS support for frontend integration.
- **Graceful Shutdown**: Handles SIGINT/SIGTERM for safe server shutdown.

## Project Structure

```
go.mod, go.sum           # Go modules
main.go                  # Entry point
clickhouse-config/       # ClickHouse config files
  users.xml

database/                # Database connection and migration scripts
  clickhouse.go
  postgres.go
  migration/
    Clickhouse.sql
    Users.sql

handlers/                # HTTP route handlers
  auth_handlers.go
  health_check.go
  track_handlers.go

middleware/              # Gin middleware (auth, CORS)
  auth_middleware.go
  cors.go

models/                  # Data models
  event.go
  user.go


store/                   # Data access layer
  analytics_store.go
  user_store.go

utils/                   # Utility functions
  helpers.go
  jwt_utils.go
  session_utils.go
```

## API Endpoints

### Public
- `POST /api/signup` — User registration
- `POST /api/login` — User login
- `POST /api/logout` — User logout

### Protected (JWT required)
- `POST /api/track` — Track an event
- `GET /api/profile` — Get user profile and IP address
- `GET /api/stats/event-counts` — Event counts over time
- `GET /api/stats/average-event-duration` — Average event duration
- `GET /api/stats/average-custom-param` — Average of a custom event parameter
- `GET /api/stats/unique-users` — Unique users over time
- `GET /api/stats/top-paths` — Top N page paths

## Setup

1. **Clone the repository**
   ```sh
   git clone <repo-url>
   cd api
   ```
2. **Configure environment variables**
   - Copy `.env.example` to `.env` and fill in your database credentials and secrets.

3. **Run database migrations**
   - Apply SQL scripts in `database/migration/` to your PostgreSQL and ClickHouse instances.

4. **Install dependencies**
   ```sh
   go mod tidy
   ```

5. **Start the server**
   ```sh
   go run main.go
   ```
   The server will start on `http://localhost:8080` by default.

## Example .env Configuration

```
PORT=8080
GIN_MODE=debug
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=your_pg_user
POSTGRES_PASSWORD=your_pg_password
POSTGRES_DB=your_pg_db
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_DB=your_ch_db
JWT_SECRET=your_jwt_secret
```

## ClickHouse Setup

1. **Install ClickHouse**
docker run -d --name clickhouse-server \
  --ulimit nofile=262144:262144 \
  -p 8123:8123 \
  -p 9000:9000 \
  clickhouse/clickhouse-server

2. **Table in ClickHouse Server**
 CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID,
    eventType String,
    userId String,
    sessionId String,
    timestamp DateTime64(3),
    pagePath String,
    durationMs UInt64,
    eventData JSON
) ENGINE = MergeTree()
ORDER BY (eventType, timestamp);

3. **Configure Users**
   - Edit `clickhouse-config/users.xml` as needed for user authentication and permissions.

4. **Create Database and Tables**
   - Run the SQL scripts in `database/migration/Clickhouse.sql` to set up the required tables:
     ```sh
     clickhouse client --query="CREATE DATABASE IF NOT EXISTS your_ch_db;"
     clickhouse client --database=your_ch_db < database/migration/Clickhouse.sql
     ```

## Environment Variables

- `PORT` — Port to run the server (default: 8080)
- `GIN_MODE` — Gin mode (`debug` or `release`)
- `POSTGRES_*` — PostgreSQL connection details
- `CLICKHOUSE_*` — ClickHouse connection details
- `JWT_SECRET` — Secret for JWT signing

## License

MIT
