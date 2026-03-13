# PINTOUR - Tour & Travel International Backend

A complete backend for **PINTOUR**, a Tour & Travel International platform, built with Go.

## Tech Stack

- **Go 1.24** – backend language
- **gRPC + Protocol Buffers** – service definitions and RPC
- **grpc-gateway** – HTTP/REST API from gRPC
- **Swagger / OpenAPI 2.0** – API documentation
- **sqlc** – type-safe database code generation
- **PostgreSQL** – relational database
- **Google OAuth2** – social login
- **Midtrans** – payment gateway (Snap)
- **JWT** – stateless authentication
- **Docker / Docker Compose** – containerization

## Services

| Service | Description |
|---------|-------------|
| `AuthService` | Register, Email/Password login, Google OAuth2 login, Profile management |
| `TourService` | Browse tour packages, destinations, schedules; CRUD for admins |
| `BookingService` | Create bookings, list history, cancel, admin status updates |
| `PaymentService` | Midtrans Snap payment, webhook handler, admin payment list |
| `ReviewService` | Create and list tour reviews; admin moderation |

## API Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/auth/register` | Register new user |
| `POST` | `/v1/auth/login` | Login with email/password |
| `POST` | `/v1/auth/google` | Login with Google OAuth2 code |
| `GET`  | `/v1/auth/profile` | Get current user profile 🔒 |
| `PUT`  | `/v1/auth/profile` | Update user profile 🔒 |

### Tours
| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/v1/tours` | List tour packages (with search & filter) |
| `GET`  | `/v1/tours/{id}` | Get tour package detail |
| `GET`  | `/v1/destinations` | List destinations |
| `GET`  | `/v1/tours/{tour_package_id}/schedules` | List available schedules |
| `POST` | `/v1/admin/tours` | Create tour package 🔒 Admin |
| `PUT`  | `/v1/admin/tours/{id}` | Update tour package 🔒 Admin |
| `DELETE` | `/v1/admin/tours/{id}` | Delete tour package 🔒 Admin |
| `POST` | `/v1/admin/destinations` | Create destination 🔒 Admin |
| `POST` | `/v1/admin/tours/{tour_package_id}/schedules` | Create schedule 🔒 Admin |

### Bookings
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/bookings` | Create booking 🔒 |
| `GET`  | `/v1/bookings/{id}` | Get booking detail 🔒 |
| `GET`  | `/v1/bookings` | List bookings 🔒 |
| `POST` | `/v1/bookings/{id}/cancel` | Cancel booking 🔒 |
| `PUT`  | `/v1/admin/bookings/{id}/status` | Update booking status 🔒 Admin |

### Payments
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/payments` | Create Midtrans Snap payment 🔒 |
| `GET`  | `/v1/payments/{booking_id}` | Get payment status 🔒 |
| `POST` | `/v1/payments/notification` | Midtrans webhook |
| `GET`  | `/v1/admin/payments` | List all payments 🔒 Admin |

### Reviews
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/reviews` | Create review 🔒 |
| `GET`  | `/v1/tours/{tour_package_id}/reviews` | List tour reviews |
| `DELETE` | `/v1/admin/reviews/{id}` | Delete review 🔒 Admin |

## Architecture — Domain-Driven Design (DDD)

This project follows the **Domain-Driven Design (DDD)** layered architecture:

```
┌──────────────────────────────────────────────────┐
│  Interface Layer  (internal/interface/)           │
│  - grpc/          thin gRPC handlers              │
│  - middleware/    auth interceptor                │
├──────────────────────────────────────────────────┤
│  Application Layer  (internal/application/)       │
│  - auth/          use-case services + DTOs        │
│  - tour/                                         │
│  - booking/                                      │
│  - payment/                                      │
│  - review/                                       │
├──────────────────────────────────────────────────┤
│  Domain Layer  (internal/domain/)                 │
│  - entity/        pure business entities         │
│  - repository/    repository interfaces (ports)  │
├──────────────────────────────────────────────────┤
│  Infrastructure Layer  (internal/infrastructure/) │
│  - persistence/   sqlc-backed repository impls   │
│  - payment/       Midtrans gateway adapter       │
│  - oauth/         Google OAuth provider          │
└──────────────────────────────────────────────────┘
```

### Dependency Rule
Each layer only depends on layers below it:
- **Interface** → Application → Domain ← Infrastructure

## Project Structure

```
pintour/
├── cmd/server/                 # Entry point — wires all DDD layers
├── internal/
│   ├── config/                 # Configuration (viper)
│   ├── token/                  # JWT token management (cross-cutting)
│   ├── util/                   # Helpers: password, random (cross-cutting)
│   ├── db/
│   │   ├── migration/          # PostgreSQL schema migrations
│   │   ├── query/              # SQL source queries (sqlc input)
│   │   └── sqlc/               # sqlc-generated type-safe DB code
│   │
│   ├── domain/                 # ── DOMAIN LAYER ──
│   │   ├── entity/             # Pure business entities (no external deps)
│   │   │   ├── user.go
│   │   │   ├── tour.go
│   │   │   ├── booking.go
│   │   │   ├── payment.go
│   │   │   └── review.go
│   │   └── repository/         # Repository interfaces (ports)
│   │       ├── user.go
│   │       ├── tour.go
│   │       ├── booking.go
│   │       ├── payment.go
│   │       └── review.go
│   │
│   ├── application/            # ── APPLICATION LAYER ──
│   │   ├── auth/               # Use-cases: Register, Login, GoogleLogin, GetProfile
│   │   ├── tour/               # Use-cases: ListPackages, GetPackage, CreatePackage…
│   │   ├── booking/            # Use-cases: CreateBooking, ListBookings, Cancel…
│   │   ├── payment/            # Use-cases: CreatePayment, HandleNotification…
│   │   └── review/             # Use-cases: CreateReview, ListReviews, Delete…
│   │
│   ├── infrastructure/         # ── INFRASTRUCTURE LAYER ──
│   │   ├── persistence/        # Repository implementations backed by sqlc
│   │   ├── payment/            # Midtrans Snap gateway adapter
│   │   └── oauth/              # Google OAuth2 provider
│   │
│   └── interface/              # ── INTERFACE LAYER ──
│       ├── grpc/               # Thin gRPC handlers (proto ↔ DTO translation only)
│       └── middleware/         # gRPC auth interceptor
│
├── pb/pintour/v1/              # Generated proto Go code
├── proto/pintour/v1/           # Proto definitions (source of truth)
├── third_party/                # googleapis + grpc-gateway protos
├── docs/swagger/               # Generated Swagger JSON
├── app.env.example             # Configuration template
├── docker-compose.yml          # PostgreSQL + server containers
└── Makefile                    # Dev commands
```

## Getting Started

### Prerequisites
- Go 1.24+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for local runs)

### 1. Configuration

```bash
cp app.env.example app.env
# Edit app.env with your credentials:
# - GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET
# - MIDTRANS_SERVER_KEY / MIDTRANS_CLIENT_KEY
# - TOKEN_SYMMETRIC_KEY (must be 32+ characters)
```

### 2. Run with Docker Compose

```bash
docker-compose up
```

This starts PostgreSQL and the PINTOUR server. The app auto-migrates the database on startup.

### 3. Run locally

```bash
# Start PostgreSQL
make postgres

# Run migrations
make migrateup

# Start the server
make server
```

### 4. API Documentation

Once running, the Swagger JSON is available at:
```
http://localhost:8080/swagger.json
```

You can load this into [Swagger UI](https://swagger.io/tools/swagger-ui/) or [Postman](https://www.postman.com/).

## Development Commands

```bash
make proto       # Regenerate pb/ from proto files
make sqlc        # Regenerate db/sqlc/ from SQL queries
make migrateup   # Apply database migrations
make migratedown # Revert database migrations
make build       # Build binary to bin/server
make test        # Run tests
```

## Authentication

All protected endpoints (`🔒`) require an `Authorization: Bearer <token>` header.

1. Register or Login → receive `access_token`
2. Include token in subsequent requests

### Google OAuth2 Flow
1. Redirect user to Google consent URL
2. Google returns an authorization `code`
3. POST `/v1/auth/google` with `{ "code": "..." }`
4. Receive JWT token

## Payment Flow (Midtrans Snap)

1. Create a booking → POST `/v1/bookings`
2. Create a payment → POST `/v1/payments` with `booking_id`
3. Redirect user to `payment_url` (Midtrans Snap)
4. Midtrans sends webhook to `/v1/payments/notification`
5. Booking status updates to `confirmed` upon successful payment
