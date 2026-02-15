# Secure Auth Service (React + Go + JWT)

## Overview

This project demonstrates a secure authentication system built using:

* React (TypeScript + Tailwind CSS)
* Go (Golang) backend
* Firebase Authentication (Identity Provider)
* JWT-based short-lived access tokens
* Opaque refresh tokens stored in PostgreSQL

The system implements secure token handling, refresh token rotation, protected API routes, logout revocation, and interceptor-based session refresh. The goal is to demonstrate production-style authentication architecture rather than simple login functionality.

---

## Repository Structure

```
root/
├── client/                 # React frontend
│   ├── src/
│   │   ├── api/
│   │   │   └── axios.ts
│   │   ├── App.tsx
│   │   ├── firebase.ts
│   │   ├── main.tsx
│   │   └── index.css
│   ├── .env
│   ├── package.json
│   └── vite.config.ts
│
├── cmd/server/             # Go backend entry point
│   └── main.go
│
├── internal/               # Backend application logic
│   ├── auth/
│   └── database/
│
├── scripts/
│   └── database.sql
│
├── firebase-service-account.json
├── go.mod
└── go.sum
```

Frontend and backend are independent applications within a single repository.

---

## Architecture

### Frontend (client/)

* Uses Firebase SDK for email/password authentication
* Retrieves Firebase ID token after login
* Exchanges ID token with backend for:

  * Access token (JWT)
  * Refresh token (httpOnly cookie)
* Stores access token in memory only
* Uses Axios interceptor to:

  * Attach Authorization header
  * Automatically call `/auth/refresh` on 401
  * Retry the original request

---

### Backend (Go)

* Verifies Firebase ID token using Firebase Admin SDK
* Issues:

  * Access token (JWT, short-lived)
  * Refresh token (UUID, long-lived)
* Stores refresh tokens in PostgreSQL
* Implements refresh token rotation
* Revokes refresh token on logout
* Protects routes using JWT middleware
* Uses CORS with credentials enabled

---

## Authentication Flow

1. User logs in using Firebase (frontend).

2. Frontend receives Firebase ID token.

3. Frontend calls:

   POST `/auth/exchange`

4. Backend:

   * Verifies Firebase token
   * Issues access token (JWT)
   * Stores refresh token in database
   * Sets refresh token in httpOnly cookie

5. Frontend stores access token in memory.

6. Protected requests use:

   Authorization: Bearer `<access_token>`

7. If access token expires:

   * Frontend receives 401
   * Automatically calls `/auth/refresh`
   * Backend validates refresh token
   * Rotates refresh token
   * Returns new access token
   * Frontend retries original request

---

## Security Features

* Access token stored in memory only
* Refresh token stored in httpOnly cookie
* Refresh token rotation implemented
* Refresh token revocation on logout
* JWT signature validation (HS256)
* CORS configured with credentials
* Rate limiting middleware for auth endpoints
* Firebase Admin credentials stored server-side only

---

## Database Schema

Run the SQL inside:

```
scripts/database.sql
```

Table:

```
refresh_tokens:
  id UUID
  user_id TEXT
  token TEXT
  expires_at TIMESTAMP
  revoked_at TIMESTAMP NULL
```

---

## Setup Instructions

### 1. Backend Setup

From repository root:

1. Install Go (1.22+ recommended)

2. Install PostgreSQL

3. Create database:

   CREATE DATABASE secure_auth;

4. Run schema from:

   scripts/database.sql

5. Place Firebase Admin credentials file:

   firebase-service-account.json

6. Run:

   go mod tidy
   go run ./cmd/server

Backend runs on:

[http://localhost:8080](http://localhost:8080)

---

### 2. Frontend Setup

Navigate to client folder:

```
cd client
```

1. Install dependencies:

   npm install

2. Configure Firebase project credentials in:

   src/firebase.ts

3. Start frontend:

   npm run dev

Frontend runs on:

[http://localhost:5173](http://localhost:5173)

---

## API Endpoints

### Authentication

POST `/auth/exchange`
Exchange Firebase ID token for access + refresh tokens.

POST `/auth/refresh`
Validate refresh token and issue new access token.

POST `/auth/logout`
Revoke refresh token and clear cookie.

---

### Protected

GET `/profile`
Requires Authorization header with valid access token.

---

## Testing Refresh Flow

1. Login.
2. Wait for access token to expire.
3. Click "Reload Profile".
4. Observe network:

   * `/profile` → 401
   * `/auth/refresh` → 200
   * `/profile` → 200

This demonstrates automatic token refresh.

---

## Why This Project Matters

This project demonstrates:

* Secure session management
* Token rotation strategy
* Backend modular architecture
* Dependency injection in Go
* Frontend interceptor-based refresh flow
* Understanding of authentication beyond basic login

It reflects patterns used in real-world production systems.

---

## Future Improvements

* Move JWT secret to environment variables
* Use RS256 with public/private keys
* Add Redis support
* Add CSRF protection
* Dockerize backend and database
* Add integration tests
* Add structured logging