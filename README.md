# RideOn — RideShare

A ride-sharing application with Go backend services and a Next.js map interface. Riders can select a destination and preview a driving route with fare estimates for SUV, sedan, van, and luxury packages.

The project currently implements trip previews. Trip booking, driver matching, and payments are still in progress.

## What works today

- **Trip previews:** the API gateway calls the trip service over gRPC; the trip service fetches a route from OSRM, estimates package fares, and stores fare quotes in memory.
- **Rider and driver screens:** a Next.js 15 / React 19 frontend uses Leaflet for maps and provides a vehicle-package selector for drivers.
- **WebSocket scaffolding:** rider and driver endpoints accept connections. Drivers receive a registration message; incoming messages are currently logged.
- **Backend validation:** coordinate checks, routing timeouts, bounded JSON bodies, HTTP error mapping, and concurrency-safe in-memory writes.

## Architecture

```text
Next.js frontend (:3000)
        |
        | HTTP / WebSocket
        v
Go API gateway (:8081)
        |
        | gRPC
        v
Go trip service (:9093) ── HTTP ──> OSRM routing API
        |
        v
In-memory trips and fare quotes
```

The trip service separates handlers, business logic, and infrastructure. Routing is injected through a `RouteProvider` interface, and persistence through `TripRepository`, so tests can replace external dependencies.

```text
services/
  api-gateway/       HTTP endpoints, WebSocket handlers, and gRPC client
  trip-service/
    cmd/             Server startup and dependency wiring
    internal/
      domain/        Trip and fare models; service/repository interfaces
      service/       Route orchestration and fare calculation
      infrastructure/  gRPC/HTTP handlers, OSRM adapter, in-memory storage
    pkg/types/       Routing response types and pricing configuration
proto/               Protobuf service definitions
shared/              Generated gRPC code, contracts, types, and utilities
web/                 Next.js frontend
infra/               Dockerfiles and Kubernetes manifests
docs/architecture/  Design diagrams for the planned booking/event flows
```

## Run locally

You need Go compatible with `go.mod` (Go 1.25 or newer), Node.js and npm. The preview flow needs access to the configured OSRM server. No database or message broker is needed for the current implementation.

Run each component in a separate terminal, starting from the repository root.

**Trip service:**

```bash
go run ./services/trip-service/cmd
```

**API gateway:**

```bash
TRIP_SERVICE_URL=localhost:9093 go run ./services/api-gateway
```

**Frontend:**

```bash
cd web
npm install
npm run dev
```

Open [localhost:3000](http://localhost:3000), choose **I Need a Ride**, and select a destination on the map to request a preview.

### Configuration

| Variable                    | Component    | Default                                                      |
| --------------------------- | ------------ | ------------------------------------------------------------ |
| `HTTP_ADDR`                 | API gateway  | `:8081`                                                      |
| `TRIP_SERVICE_URL`          | API gateway  | `trip-service:9093`; use `localhost:9093` outside Kubernetes |
| `OSRM_URL`                  | Trip service | `http://router.project-osrm.org`                             |
| `NEXT_PUBLIC_API_URL`       | Frontend     | `http://localhost:8081`                                      |
| `NEXT_PUBLIC_WEBSOCKET_URL` | Frontend     | `ws://localhost:8081/ws`                                     |

The trip service currently listens on the fixed address `:9093`. Set frontend overrides in `web/.env.local` before starting or building the frontend. The routing client has a four-second timeout; gateway RPC calls have a five-second deadline.

## API

| Endpoint                                       | Purpose                                      | Status                                                            |
| ---------------------------------------------- | -------------------------------------------- | ----------------------------------------------------------------- |
| `POST /trip/preview`                           | Return a route and package fare quotes       | Implemented                                                       |
| `POST /trip/start`                             | Start a trip using `userID` and `rideFareID` | gRPC method is unimplemented; returns HTTP 501 for valid requests |
| `GET /ws/riders?userID=...`                    | Rider WebSocket connection                   | Connection/message scaffolding                                    |
| `GET /ws/drivers?userID=...&packageSlug=sedan` | Driver WebSocket connection                  | Registration/message scaffolding                                  |

Request a preview with both backend services running:

```bash
curl -X POST http://localhost:8081/trip/preview \
  -H 'Content-Type: application/json' \
  -d '{
    "userID": "rider-demo",
    "pickUp": {"latitude": 37.7749, "longitude": -122.4194},
    "destination": {"latitude": 37.7849, "longitude": -122.4094}
  }'
```

A successful response uses a `data` envelope containing `route` and `rideFares`. Route coordinates have named `latitude` and `longitude` fields.

## Development

Run backend tests with race detection and static checks:

```bash
go test -race ./...
go vet ./...
```

Routing tests use fake HTTP transports and do not call the public OSRM service.

After changing `proto/trip.proto`, regenerate the checked-in Go bindings with `make generate-proto`. This requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on your `PATH`.

For Kubernetes development, start a local cluster and run `tilt up` with Docker, kubectl, and Tilt installed. The Tiltfile builds the gateway, trip service, and frontend, and forwards ports 8081 and 3000.

## Current limitations

- Fare quotes and trips are process-local and disappear on restart.
- User IDs are supplied by clients; authentication is not implemented.
- Driver matching, RabbitMQ event processing, and payment services are not implemented. The [architecture diagrams](docs/architecture/) describe planned flows.
- Production manifests are templates that need updating before use: they contain project placeholders, and the production trip-service port does not match the current gRPC listener.
- Kubernetes gateway manifests currently set `GATEWAY_HTTP_ADDR`, while the application reads `HTTP_ADDR`. The default port works, but custom addresses require aligning that variable.
