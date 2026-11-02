# Movie Ticket Booking System

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white)

A movie ticket booking backend built as independent Go microservices behind a single API gateway. Users browse movies and showtimes, temporarily hold seats, and confirm bookings. Administrators manage movies, auditoriums, showtimes, and auditorium seats.

## Architecture

```mermaid
flowchart LR
    Client[Client]
    Gateway[API Gateway]
    Auth[Auth Service]
    Movie[Movie Service]
    Ticket[Ticket Service]
    AuthDB[(PostgreSQL<br/>Auth tables)]
    MovieDB[(PostgreSQL<br/>Movie tables)]
    TicketDB[(PostgreSQL<br/>Booking tables)]
    Redis[(Redis<br/>Temporary seat holds)]

    Client -->|REST + JWT| Gateway
    Gateway --> Auth
    Gateway --> Movie
    Gateway --> Ticket
    Auth --> AuthDB
    Movie --> MovieDB
    Ticket --> TicketDB
    Ticket -->|Showtime and seat lookups| Movie
    Ticket --> Redis
```

The gateway validates JWTs on protected routes and forwards the authenticated user ID and role to downstream services. The ticket service calls the movie service to verify showtimes, seats, and prices.

## Booking Flow

```mermaid
sequenceDiagram
    actor User
    participant Gateway
    participant Ticket as Ticket Service
    participant Movie as Movie Service
    participant Redis
    participant DB as Ticket PostgreSQL

    User->>Gateway: POST /showtimes/:id/holds
    Gateway->>Gateway: Validate JWT and set trusted identity headers
    Gateway->>Ticket: Forward hold request
    Ticket->>Movie: Verify showtime and auditorium seats
    Ticket->>Redis: Atomically hold requested seats with a TTL
    Redis-->>Ticket: Hold ID and expiry
    Ticket-->>User: Hold details

    User->>Gateway: POST /holds/:id/confirm
    Gateway->>Ticket: Forward authenticated confirmation
    Ticket->>Redis: Verify hold owner and held seats
    Ticket->>Movie: Read showtime price
    Ticket->>DB: Insert booking and booking seats in a transaction
    Ticket->>Redis: Release hold keys
    Ticket-->>User: Confirmed booking
```

Seat holds expire after up to five minutes, capped at the showtime start. Redis Lua scripts reserve and release a batch atomically. PostgreSQL stores confirmed bookings and enforces uniqueness for active `(showtime_id, seat_id)` pairs.

## Project Structure

Each service has its own Go module, configuration, handlers, and database access code.

```text
movie-ticket-booking-system/
|-- api-gateway/
|   |-- cmd/server/
|   |-- internal/
|   `-- .env.example
|-- auth-service/
|   |-- cmd/server/
|   |-- internal/
|   |-- migrations/
|   `-- .env.example
|-- movie-service/
|   |-- cmd/server/
|   |-- internal/
|   |-- migrations/
|   `-- .env.example
|-- ticket-service/
|   |-- cmd/server/
|   |-- internal/
|   |-- migrations/
|   `-- .env.example
|-- tests/
|   `-- e2e/
|       |-- booking_test.go
|       `-- go.mod
|-- LICENSE
`-- README.md
```

## Service Ports

| Service | Port | Responsibility |
| --- | ---: | --- |
| `api-gateway` | `8080` | Client entry point, JWT validation, routing, and rate limiting |
| `auth-service` | `8081` | Registration, login, password hashing, and JWT creation |
| `movie-service` | `8082` | Movies, auditoriums, showtimes, and seats |
| `ticket-service` | `8083` | Seat holds, booking confirmation, and booking lookup |

## Services

### Auth Service

Responsible for:

* User registration and login
* Password hashing with bcrypt
* JWT generation with user ID, email, and role claims
* Default `user` role for new accounts

### Movie Service

Responsible for:

* Movie creation, browsing, and updates
* Auditorium creation
* Showtime scheduling with auditorium overlap checks
* Auditorium seat creation and showtime seat lookup
* Admin-only movie, auditorium, showtime, and seat management

### Ticket Service

Responsible for:

* Temporary multi-seat holds in Redis
* Hold ownership and expiry checks
* Booking confirmation in a PostgreSQL transaction
* Booking and booked-seat lookup
* Per-seat price snapshots at confirmation time

### API Gateway

Responsible for:

* Single entry point for client traffic
* Routing requests to auth, movie, and ticket services
* JWT verification and identity propagation
* Redis-backed rate limiting for protected routes (100 requests per user per minute)

## Tech Stack

### Backend

* Go 1.26+
* Gin
* `pgx` PostgreSQL driver and connection pool

### Data and concurrency

* PostgreSQL for users, movies, auditoriums, showtimes, and bookings
* Redis for short-lived seat holds and gateway rate limiting
* Redis Lua scripts for atomic batch hold and release operations

### Authentication

* JWT signed with HS256
* bcrypt password hashing
* Role-based checks for user and admin routes

### Architecture

* REST APIs
* API gateway and independent service modules
* Synchronous HTTP calls from ticket service to movie service

## Authentication

Register and log in through the gateway:

```http
POST /auth/register
POST /auth/login
```

Protected endpoints require:

```http
Authorization: Bearer <jwt_token>
```

The gateway validates the token and overwrites incoming identity headers with the verified claims before forwarding:

```http
X-User-ID: <user UUID>
X-User-Role: user | admin
```

Newly registered accounts have the `user` role. There is no public admin-registration endpoint; for local development, an administrator must be provisioned directly in the auth database.

## Getting Started

### 1. Prerequisites

* Go 1.26 or later
* PostgreSQL
* Redis

### 2. Create a PostgreSQL database and apply migrations

The example environment files point to a local database named `database`. Create that database or update each service's `DATABASE_URL` to use an existing one. Apply the migrations in order:

```bash
psql "$DATABASE_URL" -f auth-service/migrations/001_init.sql

psql "$DATABASE_URL" -f movie-service/migrations/001_init.sql
psql "$DATABASE_URL" -f movie-service/migrations/002_create_showtimes_and_auditoriums.sql
psql "$DATABASE_URL" -f movie-service/migrations/003_seats.sql

psql "$DATABASE_URL" -f ticket-service/migrations/001_init.sql
```

For a simple local setup, the same database URL can be used by all three services. Their migrations create separate tables. You can point the services at separate databases by setting each `DATABASE_URL` independently.

### 3. Configure environment variables

Copy each `.env.example` to `.env` in the corresponding service directory, then replace the placeholders. Each service loads `.env` from its current working directory.

**`auth-service/.env`**

```dotenv
DATABASE_URL=postgres://user:password@127.0.0.1:5432/database?sslmode=disable
JWT_SECRET=<long-random-secret>
```

**`movie-service/.env`**

```dotenv
DATABASE_URL=postgres://user:password@127.0.0.1:5432/database?sslmode=disable
```

**`ticket-service/.env`**

```dotenv
DATABASE_URL=postgres://user:password@127.0.0.1:5432/database?sslmode=disable
REDIS_URL=redis://127.0.0.1:6379/0
MOVIE_SERVICE_URL=http://127.0.0.1:8082
```

**`api-gateway/.env`**

```dotenv
JWT_SECRET=<same-long-random-secret-as-auth-service>
```

The `JWT_SECRET` value must match in auth service and API gateway. `.env` files are ignored by Git; do not commit real credentials.

### 4. Start the services

Run each command from the repository root in a separate terminal. PostgreSQL and Redis should already be running.

```bash
cd auth-service && go run ./cmd/server
```

```bash
cd movie-service && go run ./cmd/server
```

```bash
cd ticket-service && go run ./cmd/server
```

```bash
cd api-gateway && go run ./cmd/server
```

The API gateway listens at `http://localhost:8080`. Its upstream service URLs are currently configured in `api-gateway/cmd/server/main.go` for local development.

### 5. Provision an administrator for local development

Register a user through `/auth/register`, then update that account's role in the auth database and log in again to receive a token with the admin role:

```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

Do not expose direct role updates as a public API.

## API Endpoints

All client requests should go through `http://localhost:8080`.

### Public endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/auth/register` | Create a user account |
| `POST` | `/auth/login` | Authenticate and receive a JWT |
| `GET` | `/movies` | List movies |
| `GET` | `/movies/:id` | Get a movie |
| `GET` | `/movies/:id/showtimes` | List a movie's showtimes |
| `GET` | `/showtimes/:id` | Get showtime details |
| `GET` | `/showtimes/:id/seats` | List seats in the showtime auditorium |

### Protected admin endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/movies` | Create a movie |
| `PUT` | `/movies/:id` | Update a movie |
| `POST` | `/auditoriums` | Create an auditorium |
| `POST` | `/movies/:id/showtimes` | Schedule a showtime |
| `POST` | `/auditoriums/:id/seats` | Add auditorium seats in a batch |

### Protected user endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/showtimes/:id/holds` | Temporarily hold one or more seats |
| `DELETE` | `/holds/:id` | Release a hold |
| `POST` | `/holds/:id/confirm` | Confirm a hold and create a booking |
| `GET` | `/bookings/:id` | Get one of the authenticated user's bookings |
| `GET` | `/users/me/bookings` | List the authenticated user's bookings |

To create a hold, send seat IDs:

```json
{
  "seat_ids": ["<seat-uuid>"]
}
```

The response includes a `hold_id` and `expires_at`. Use that hold ID in the confirmation or release endpoint.

## Seat Hold Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Available
    Available --> Held: User creates hold
    Held --> Booked: Hold confirmed
    Held --> Available: User releases hold
    Held --> Available: Hold expires
    Booked --> [*]
```

## Build

Each service is a separate Go module. From a service directory, build it with:

```bash
go build ./...
```

## Tests

Run the unit tests from the repository root:

```bash
go -C auth-service test ./...
go -C movie-service test ./...
go -C ticket-service test ./...
go -C api-gateway test ./...
```

The unit tests cover invalid registration, login, movie, showtime, seat, and hold inputs, plus gateway JWT rejection and trusted identity-header behavior.

The opt-in end-to-end test creates temporary admin and user accounts, then creates a movie, auditorium, seat, and showtime through the gateway. It checks rejected registration and login requests, unauthenticated admin access, an invalid movie ID, an overlapping showtime, and a second user's attempt to hold an already-held seat. It also verifies the successful hold, booking confirmation, and booking lookup flow.

Start PostgreSQL, Redis, and all four services, then run the E2E test with a disposable PostgreSQL database that has all migrations applied:

```bash
cd tests/e2e
E2E_GATEWAY_URL=http://localhost:8080 \
E2E_DATABASE_URL='postgres://user:password@127.0.0.1:5432/test_database?sslmode=disable' \
go test -v
```

The E2E test skips when these variables are missing. It creates temporary user accounts (promoting one directly in the test database), runs the booking flow, and removes created database records during cleanup. Use a dedicated test database: the cleanup connects directly to PostgreSQL and deletes the rows created by the test.

## Security Notes

* Only the API gateway should be reachable by external clients.
* Downstream services trust the `X-User-ID` and `X-User-Role` headers set by the gateway; they do not independently validate JWTs. Keep service ports private or restrict them to the internal network.
* Use a strong, private JWT secret and keep the auth service and gateway values synchronized.
* The gateway's upstream URLs currently target `localhost`; update them for container or remote deployments.
* This project is under active development and has not had an independent security review.

## Future Improvements

* Add payment processing with payment status tracking and refund handling.
* Define a booking cancellation policy and implement cancellation for confirmed bookings.
* Add pagination and filtering for movie, showtime, and booking lists.
* Make service URLs configurable and add containerized deployment configuration.
* Add health checks, metrics, and structured logs for service monitoring.
* Expand automated coverage with repository tests and Redis concurrency tests.

## Current Scope

The repository contains backend services and an API gateway. It does not currently include a frontend, payment integration, booking cancellation endpoint, or deployment manifests. Seat holds and booking confirmation are implemented; cancellations and payment flows can be added according to the desired booking policy.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
