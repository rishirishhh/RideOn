# trip service

This service handles all trip-related operations in the system.

## Architecture

The service follows Clean Architecture principles with the following structure:

```
services/trip-service/
├── cmd/                    # Application entry points
│   └── main.go            # Main application setup
├── internal/              # Private application code
│   ├── domain/           # Business domain models and interfaces
│   ├── service/          # Business logic implementation
│   │   └── service.go    # Service implementations
│   └── infrastructure/   # External dependencies implementations (abstractions)
│       ├── events/       # Event handling (RabbitMQ)
│       ├── grpc/         # gRPC server handlers
│       └── repository/   # Data persistence
├── pkg/                  # Public packages
│   └── types/           # Shared types and models
└── README.md            # This file
```

### Layer Responsibilities

1. **Domain Layer** (`internal/domain/`)
   - Contains business domain interfaces
   - Defines contracts for repositories and services
   - Pure business logic, no implementation details

2. **Service Layer** (`internal/service/`)
   - Implements business logic
   - Uses repository interfaces
   - Coordinates between different parts of the system

3. **Infrastructure Layer** (`internal/infrastructure/`)
   - `repository/`: Implements data persistence
   - `events/`: Handles event publishing and consuming
   - `grpc/`: Handles gRPC communication

4. **Public Types** (`pkg/types/`)
   - Contains shared types and models
   - Can be imported by other services

## Key Benefits

1. **Dependency Inversion**: Services depend on interfaces, not implementations
2. **Separation of Concerns**: Each layer has a specific responsibility
3. **Testability**: Easy to mock dependencies for testing
4. **Maintainability**: Clear boundaries between components
5. **Flexibility**: Easy to swap implementations without affecting business logic


## Routing configuration and tests

The service receives a `RouteProvider` through its constructor. The OSRM adapter
owns HTTP requests, response size limits, and response validation. `OSRM_URL`
configures its base URL (default: `http://router.project-osrm.org`); the client
has a four-second timeout and propagates request cancellation.

OSRM geometry uses `[longitude, latitude]`. The adapter's response conversion
maps these to named protobuf fields; the web map consumes `[latitude, longitude]`.
Invalid coordinates and empty or malformed routes return errors before pricing.

The in-memory repository serializes writes and copies stored objects. It remains
process-local storage: restarting the service loses trips and fare quotes.

From the repository root, run `go test -race ./...` and `go vet ./...`.
Tests use fake routing transports and do not call the public OSRM service.
