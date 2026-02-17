# Secure Auth Service

A full-stack authentication system built with React and Go implementing JWT-based access control, refresh token rotation, secure cookie handling, configurable CORS, and IP-based rate limiting.

This project demonstrates production-style authentication architecture rather than basic login functionality.

---

## Tech Stack

### Frontend

* React (TypeScript)
* Vite
* Tailwind CSS
* Axios with interceptor-based token refresh
* Firebase Authentication (email/password)

### Backend

* Go (Golang)
* Chi Router
* PostgreSQL
* JWT (HS256)
* Firebase Admin SDK
* godotenv (environment configuration)
* Custom IP-based rate limiting middleware

---

## Repository Structure

```
secure-auth-service/
├── client/                  # React frontend
│   ├── src/
│   │   ├── api/axios.ts
│   │   ├── App.tsx
│   │   ├── firebase.ts
│   │   └── main.tsx
│   ├── .env
│   ├── package.json
│   └── vite.config.ts
│
├── cmd/server/              # Go backend entry point
│   └── main.go
│
├── internal/
│   ├── auth/                # Handlers, service, middleware
│   ├── database/            # PostgreSQL connection
│   └── middleware/          # Rate limiter
│
├── scripts/
│   └── database.sql
│
├── firebase-service-account.json
├── .env                     # Backend environment configuration
├── go.mod
└── go.sum
```

Frontend and backend are separate applications inside one repository.

---

## Authentication Architecture

### Access Token

* JWT (HS256)
* Short-lived (default 15 minutes)
* Stored in memory only
* Sent via `Authorization: Bearer <token>`

### Refresh Token

* Opaque UUID
* Stored in PostgreSQL
* Stored in httpOnly cookie
* Rotated on every refresh
* Revoked on logout

---

## Authentication Flow

1. User logs in via Firebase.

2. Frontend receives Firebase ID token.

3. Frontend calls:

   POST `/auth/exchange`

4. Backend:

   * Verifies Firebase ID token
   * Issues JWT access token
   * Stores refresh token in database
   * Sets refresh token as httpOnly cookie

5. Frontend stores access token in memory.

6. Protected requests include Authorization header.

7. If access token expires:

   * Frontend receives 401
   * Automatically calls `/auth/refresh`
   * Backend validates and rotates refresh token
   * Issues new access token
   * Frontend retries original request

---

## Security Features

* Access token stored in memory (not localStorage)
* Refresh token stored in httpOnly cookie
* Refresh token rotation implemented
* Refresh token revocation on logout
* JWT signature verification (HS256)
* Environment-based secret configuration
* Configurable CORS origins via environment variables
* IP-based rate limiting on authentication endpoints
* No secrets committed to repository

---

## Rate Limiting

Authentication endpoints (`/auth/exchange`, `/auth/refresh`) are protected by a custom in-memory rate limiter:

* IP-based tracking
* Configurable request limit per time window
* Thread-safe implementation using mutex
* Automatic cleanup of stale clients
* Returns HTTP 429 when limit is exceeded

This prevents brute-force and abuse attempts on login endpoints.

---

## Environment Configuration

### Backend (.env)

Create `.env` in repository root:

```
DATABASE_URL=postgres://postgres:admin@localhost:5433/secure_auth
JWT_SECRET=super-secret-key
ALLOWED_ORIGINS=http://localhost:5173
```

Multiple origins supported:

```
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

Variables:

* `DATABASE_URL` – PostgreSQL connection string
* `JWT_SECRET` – Secret for signing JWT tokens
* `ALLOWED_ORIGINS` – Allowed CORS origins

---

### Frontend (client/.env)

Create `.env` inside `client/`:

```
VITE_API_BASE_URL=http://localhost:8080
VITE_FIREBASE_API_KEY=your_api_key
VITE_FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your_project_id
VITE_FIREBASE_STORAGE_BUCKET=your_project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=your_sender_id
VITE_FIREBASE_APP_ID=your_app_id
```

Important:

* All frontend variables must start with `VITE_`
* Firebase client config is safe for frontend exposure
* Backend secrets must never be placed in frontend `.env`

---

## Database Schema

Run:

```
scripts/database.sql
```

Table:

```
refresh_tokens:
  id UUID PRIMARY KEY
  user_id TEXT
  token TEXT
  expires_at TIMESTAMP
  revoked_at TIMESTAMP NULL
```

---

## Setup Instructions

### Backend

From repository root:

1. Install Go (1.22+ recommended)

2. Install PostgreSQL

3. Create database:

   CREATE DATABASE secure_auth;

4. Run schema from `scripts/database.sql`

5. Add `firebase-service-account.json`

6. Create `.env`

7. Run:

   go mod tidy
   go run ./cmd/server

Backend runs on:

[http://localhost:8080](http://localhost:8080)

---

### Frontend

```
cd client
npm install
npm run dev
```

Frontend runs on:

[http://localhost:5173](http://localhost:5173)

---

## API Endpoints

### Authentication

POST `/auth/exchange`
Exchange Firebase ID token for access + refresh tokens.

POST `/auth/refresh`
Validate refresh token, rotate refresh token, issue new access token.

POST `/auth/logout`
Revoke refresh token and clear cookie.

---

### Protected

GET `/profile`
Requires valid JWT access token in Authorization header.

---

## Unit Tests

Includes basic tests for:

* JWT creation and validation
* Refresh token rotation logic

Run:

```
go test ./...
```

---

## Why This Project Matters

This project demonstrates:

* Secure session management
* Token rotation strategy
* Clean modular backend architecture
* Dependency injection patterns in Go
* Environment-based configuration
* Interceptor-driven automatic session renewal
* Rate limiting for auth endpoints
* Production-style authentication design