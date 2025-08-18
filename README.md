# User Service

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A microservice responsible for user identity and authentication management within the E-Commerce Microservice Platform.

## 📖 Overview

This service is the central authority for all user-related operations. It handles the core logic for:

*   **User Registration:** Creating new user accounts with securely hashed passwords.
*   **User Authentication:** Validating user credentials (email and password).
*   **Token Generation:** Issuing JWT (JSON Web Tokens) upon successful authentication, which are used by the API Gateway to authorize subsequent requests.

It communicates with a PostgreSQL database to persist user data and exposes its functionality through a gRPC API.

## 🚀 Architecture

The User Service follows a standard Clean Architecture pattern, separating concerns into distinct layers:

*   **Handler:** Receives incoming gRPC requests, validates them, and calls the appropriate service method.
*   **Service:** Contains the core business logic, such as password hashing/comparison and user validation.
*   **Repository:** Handles all direct database interactions, abstracting the data storage from the business logic.

```mermaid
graph TD
    GW[🌐 API Gateway] -- gRPC --> Handler;

    subgraph User Service
        Handler -- Calls --> BusinessService[Service];
        BusinessService -- Uses --> Repository;
        Repository -- CRUD --> DB[(🗄️ PostgreSQL)];
    end

    subgraph Observability Stack
      Handler -- Traces --> JAEGER[(🔍 Jaeger)];
      Repository -- Traces --> JAEGER;
      Handler -- Metrics --> PROM[(📊 Prometheus)];
      Handler -- Logs --> EFK[(📝 EFK Stack)];
    end
```

### Security

*   **Password Hashing:** Passwords are never stored in plain text. The service uses the robust **bcrypt** algorithm to generate secure hashes of user passwords before saving them to the database.
*   **JWT Generation:** On successful login, the service creates a JWT containing the `user_id` and `role` claims, signed with a secret key.

## 🛠️ Tech Stack

| Category | Technology |
| :--- | :--- |
| **Language** | Go |
| **Communication** | gRPC (Server), Protobuf |
| **Database** | PostgreSQL (`pgx/v5` driver) |
| **Security** | `golang.org/x/crypto/bcrypt`, `golang-jwt/jwt` |
| **Observability** | OpenTelemetry (`otelgrpc`, `otelpgx`), Prometheus, Zap |
| **Containerization** | Docker |

## ⚙️ Setup & Running

This service is designed to be run as part of the complete platform using Docker Compose.

### Prerequisites

*   Docker & Docker Compose
*   All project repositories must be cloned at the same directory level.

### Running

The service is automatically built and started when you run the main `docker-compose up` command from the `docker-compose-environment` repository.

For detailed instructions, please see the [main project README](https://github.com/ogozo/docker-compose-environment/blob/main/README.md).

### Environment Variables

The service is configured via the `env/user.env` file in the main compose directory.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `GRPC_PORT` | The port on which the gRPC server listens. | `:50051` |
| `METRICS_PORT`| The port for the `/metrics` HTTP endpoint. | `:9001` |
| `DATABASE_URL`| Connection string for the PostgreSQL database. | `postgres://admin:secret@postgres-db:5432/ecommerce`|
| `JWT_SECRET_KEY`| Secret key for signing JWTs. | `super-secret-key` |
| `OTEL_SERVICE_NAME`| Service name for observability. | `service-user` |
| `OTEL_EXPORTER_OTLP_ENDPOINT`| OTLP endpoint for Jaeger. | `jaeger-tracer:4317` |

## 📡 gRPC API

This service exposes its functionality via the `UserService` gRPC interface, defined in the `proto-definitions` repository.

### `rpc Register(RegisterRequest) returns (RegisterResponse)`
*   **Description:** Creates a new user in the database.
*   **Request:** `RegisterRequest { string email, string password }`
*   **Response:** `RegisterResponse { User user }`

### `rpc Login(LoginRequest) returns (LoginResponse)`
*   **Description:** Authenticates a user and returns a JWT.
*   **Request:** `LoginRequest { string email, string password }`
*   **Response:** `LoginResponse { string access_token }`
