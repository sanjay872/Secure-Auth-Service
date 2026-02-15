# Secure Auth Service

A full-stack authentication system built with React and Go implementing JWT-based access control, refresh token rotation, secure cookie handling, and interceptor-driven session renewal.

This project demonstrates production-style authentication architecture rather than basic login functionality.

---

## Tech Stack

**Frontend**

* React (TypeScript)
* Vite
* Tailwind CSS
* Axios (interceptors)
* Firebase Authentication (email/password)

**Backend**

* Go (Golang)
* Chi Router
* PostgreSQL
* JWT (HS256)
* Firebase Admin SDK
* godotenv (environment configuration)

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
│   ├── auth/                # Auth handlers, middleware, service
│   └── database/            # PostgreSQL connection
│
├── scripts/
│   └── database.sql         # DB schema
│
├── firebase-service-account.json
├── .env                     # Backend environment config
├── go.mod
└── go.sum
```

Frontend and backend are independent applications within a single repository.

---

## Authentication Architecture

### Access Token

* JWT (HS256)
* Short-lived
* Stored in memory (never in localStorage)
* Sent in `Authorization: Bearer` header

### Refresh Token

* Opaque UUID
* Stored in PostgreSQL
* Stored in httpOnly cookie
* Rotated on every refresh
* Revoked on logout

---

## Authentication Flow

1. User logs in via Firebase (frontend).

2. Firebase returns an ID token.

3. Frontend calls:

   POST `/auth/exchange`

4. Backend:

   * Verifies Firebase ID token
   * Issues access token
   * Stores refresh token in database
   * Sets refresh token in httpOnly cookie

5. Frontend stores access token in memory.

6. Protected requests include Authorization header.

7. If access token expires:

   * Frontend receives 401
   * Automatically calls `/auth/refresh`
   * Backend validates and rotates refresh token
   * Returns new access token
   * Frontend retries original request

---

## Security Features

* Access token stored in memory only
* Refresh token stored in httpOnly cookie
* Refresh token rotation implemented
* Refresh token revocation on logout
* JWT signature verification (HS256)
* Environment-based secret configuration
* Configurable CORS origins
* Basic rate limiting middleware
* No secrets committed to repository

---

Good catch — yes, the frontend `.env` absolutely needs to be documented. Especially since you're using Firebase and possibly a backend base URL.

Here is the updated README section with both **backend and frontend environment configuration** clearly explained.

You can replace the **Environment Configuration** section with this.

---

## Environment Configuration

### Backend (.env)

Backend configuration is managed using environment variables loaded via `godotenv`.

Create a `.env` file in the root of the repository:

```
DATABASE_URL=postgres://postgres:admin@localhost:5433/secure_auth
JWT_SECRET=super-secret-key
ALLOWED_ORIGINS=http://localhost:5173
```

Multiple origins are supported using comma-separated values:

```
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

**Variables**

* `DATABASE_URL` – PostgreSQL connection string
* `JWT_SECRET` – Secret used to sign JWT access tokens
* `ALLOWED_ORIGINS` – Comma-separated list of allowed frontend origins for CORS

Secrets are never hardcoded in the backend.

---

### Frontend (.env)

The frontend also uses environment variables via Vite.

Create a `.env` file inside the `client/` directory:

```
VITE_API_BASE_URL=http://localhost:8080
VITE_FIREBASE_API_KEY=your_api_key
VITE_FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your_project_id
VITE_FIREBASE_APP_ID=your_app_id
```

**Important**

* All frontend environment variables must be prefixed with `VITE_`
* These values are safe for frontend exposure (Firebase client config is public by design)
* Backend secrets must never be placed in frontend `.env`

---

## How Environment Variables Are Used

### Backend

* `DATABASE_URL` initializes PostgreSQL connection
* `JWT_SECRET` signs and verifies access tokens
* `ALLOWED_ORIGINS` configures CORS dynamically

### Frontend

* `VITE_API_BASE_URL` sets Axios base URL
* Firebase configuration initializes client authentication SDK

---

## Security Notes

* Backend secrets are server-side only.
* Frontend Firebase configuration is public and not considered sensitive.
* Refresh tokens are stored in httpOnly cookies.
* Access tokens are stored in memory only.

---

## Database Schema

Run:

```
scripts/database.sql
```

Table structure:

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
Validate refresh token and issue new access token (rotates refresh token).

POST `/auth/logout`
Revoke refresh token and clear cookie.

---

### Protected

GET `/profile`
Requires valid access token in Authorization header.

---

## Testing Session Refresh

1. Login.
2. Wait for access token to expire.
3. Click “Reload Profile”.
4. Observe network:

   * `/profile` → 401
   * `/auth/refresh` → 200
   * `/profile` → 200

Demonstrates automatic session renewal.

---

## Why This Project Is Important

This project demonstrates:

* Secure session management
* Token rotation strategy
* Separation of concerns in Go backend
* Dependency injection patterns
* Environment-based configuration
* Interceptor-driven token refresh on frontend
* Real-world authentication architecture

It reflects patterns commonly used in production systems.

---

## Future Improvements

* Use RS256 with key pairs
* Add Redis support
* Add CSRF protection
* Add structured logging
* Add Docker Compose setup
* Add integration tests
* Add role-based access control

---