# OPP Backend

The OPP Backend serves as the core API component of the Open Park Project, providing endpoints for ticket management, car registration, and user management.

## Getting Started

### Development Mode

```bash
# Build and run with Docker Compose (auth service bypassed)
docker-compose build
docker-compose up
```

When running in development mode:
- API is available at http://localhost:8080
- Auth service is bypassed
- Set `DEBUG_MODE=true` environment variable to bypass authentication (already set in `docker-compose.yml`)

### Production Mode

```bash
# Run the published container image
docker run -p 8080:8080 ghcr.io/openparkproject/opp-backend:latest
```

## Authentication Flow

The authentication system follows a modern, secure pattern:

```
Client → Nginx → Auth Service → JWT Token → Backend API
```

1. **Authentication Service**: A separate service handles user authentication
   - Processes login/registration requests via `/session` endpoints
   - Issues JWT tokens signed with a private key
   - Exposes a public key for token verification

2. **Backend Authentication**:
   - Validates JWT tokens passed as Bearer tokens in the Authorization header
   - Retrieves the public key from the auth service for verification
   - Enforces proper authorization based on JWT claims (permissions/roles)
