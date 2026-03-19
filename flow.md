# Go Training — Project Summary

A step-by-step REST API project covering Go fundamentals through advanced topics.

## Project Architecture

```
Client Request
  → Recoverer Middleware (catches panics)
    → Logger Middleware (logs method, path, status, duration)
      → Auth Middleware (validates X-API-Key header, sets context)
        → Handler (parses request, calls store)
          → UserStore Interface
            → SQLiteStore (production) OR MockStore (tests)
              → SQLite Database (persistent storage)
```

## File Structure

```
go-training/
├── main.go              # Server startup, routing, graceful shutdown
├── config.go            # Environment variable config loader
├── model.go             # User struct + UserStore interface
├── handler.go           # HTTP handlers (CRUD operations)
├── database.go          # SQLiteStore — implements UserStore
├── middleware.go         # Logger + Recoverer middleware
├── auth.go              # API Key auth middleware + context helpers
├── apperror.go          # Custom error type with HTTP status codes
├── concurrency.go       # Goroutine, channel, worker pool examples
├── generics.go          # Generic functions, constraints, structs
├── *_test.go            # Tests for every module
├── Dockerfile           # Multi-stage build
├── .dockerignore        # Files excluded from Docker build
├── Jenkinsfile          # CI pipeline
└── deployment/
    ├── deployment.yaml  # K8s Deployment with probes + resources
    ├── service.yaml     # K8s Service (ClusterIP)
    ├── configmap.yaml   # K8s ConfigMap (env vars)
    └── pvc.yaml         # K8s PersistentVolumeClaim (SQLite storage)
```

## Step-by-Step Summary

### Step 1 — GET /users/{id}

| Concept | What it does |
|---------|-------------|
| `strings.TrimPrefix` | Extracts ID from URL path manually |
| `strconv.Atoi` | Converts string to int (`"3"` → `3`) |
| `for _, item := range slice` | Iterates over a slice; `_` ignores the index |
| `json.NewEncoder(w).Encode(v)` | Writes a struct as JSON to the HTTP response |
| HTTP 404 vs 400 | 404 = resource not found, 400 = bad request (invalid input) |

### Step 2 — PUT /users/{id}

| Concept | What it does |
|---------|-------------|
| `for i := range slice` | Iterates with index — needed to modify the original slice |
| Value vs pointer semantics | `for _, v := range` gives a copy; `slice[i]` gives the original |
| `json.NewDecoder(r.Body).Decode(&v)` | Reads JSON from request body into a struct via pointer |
| Partial update | Only overwrite fields that have non-zero values |

### Step 3 — Chi Router

| Concept | What it does |
|---------|-------------|
| `go get` | Downloads third-party packages; updates `go.mod` and `go.sum` |
| `chi.NewRouter()` | Creates a router that implements `http.Handler` |
| `r.Get`, `r.Post`, `r.Put`, `r.Delete` | Registers routes per HTTP method (no more switch statements) |
| `{id}` URL parameter | Chi extracts it automatically; read with `chi.URLParam(r, "id")` |
| `chi.RouteContext` in tests | Required for `chi.URLParam` to work in test environments |

### Step 4 — SQLite Database

| Concept | What it does |
|---------|-------------|
| `_ "github.com/mattn/go-sqlite3"` | Blank import — driver registers itself via `init()` |
| `sql.Open("sqlite3", path)` | Creates a connection pool (doesn't connect yet) |
| `db.Exec(query, args...)` | Runs INSERT, UPDATE, DELETE (no rows returned) |
| `db.Query(query, args...)` | Runs SELECT, returns multiple rows |
| `db.QueryRow(query, args...)` | Returns a single row (0 or 1) |
| `rows.Scan(&v1, &v2, ...)` | Reads column values into variables via pointers |
| `defer rows.Close()` | Ensures rows are closed when the function exits |
| `?` placeholder | Prevents SQL injection — never concatenate user input into SQL |
| `":memory:"` | In-memory SQLite for tests — destroyed when connection closes |
| `result.LastInsertId()` | Gets the auto-generated ID after an INSERT |

### Step 5 — Middleware (Logger & Recoverer)

| Concept | What it does |
|---------|-------------|
| Middleware pattern | `func(next http.Handler) http.Handler` — wraps a handler |
| `r.Use(middleware)` | Applies middleware to all routes on the router |
| `http.HandlerFunc` adapter | Converts a function into an `http.Handler` |
| Struct embedding | `struct { http.ResponseWriter }` — inherits all methods |
| `time.Since(start)` | Calculates elapsed time for performance logging |
| `defer + recover()` | Catches panics — Go's equivalent of try/catch |

### Step 6 — Auth Middleware (API Key & Context)

| Concept | What it does |
|---------|-------------|
| `r.Header.Get("X-API-Key")` | Reads a value from HTTP request headers |
| `map[string]string` | Go's hash map — key-value pairs |
| `value, ok := map[key]` | Comma-ok idiom — checks if a key exists |
| HTTP 401 vs 403 | 401 = identity unknown (missing key), 403 = identity known but forbidden |
| `context.WithValue(ctx, k, v)` | Adds a value to context (returns a new context — immutable) |
| `r.WithContext(ctx)` | Attaches the updated context to the request |
| `ctx.Value(key).(string)` | Reads from context + type assertion |
| `type contextKey string` | Custom type for context keys — prevents cross-package collisions |
| `r.Group(func(r chi.Router))` | Groups routes with shared middleware |

### Step 7 — Interface & Dependency Injection

| Concept | What it does |
|---------|-------------|
| `interface` | Defines method signatures — "what", not "how" |
| Implicit implementation | No `implements` keyword — any type with matching methods qualifies |
| Dependency injection | Pass dependencies into a struct: `NewHandler(store UserStore)` |
| `type Handler struct` | Handlers become struct methods — can carry state (the store) |
| `(h *Handler)` receiver | Declares a method on the Handler struct |
| `MockStore` | Fake implementation for tests — no database needed |
| Constructor function | `NewSQLiteStore()`, `NewHandler()` — Go's struct creation pattern |

### Step 8 — Error Handling

| Concept | What it does |
|---------|-------------|
| Custom error type | `AppError` struct with `Code` + `Message` fields |
| `error` interface | Built-in: any type with `Error() string` is an error |
| Type assertion | `err.(*AppError)` — checks if an error is a specific type |
| `errors.As(err, &target)` | Safely extracts a specific error type from the chain |
| `writeError(w, err)` | Central error handler — writes status code + JSON from AppError |
| Table-driven tests | `[]struct{...}` test cases + `t.Run` sub-tests |

### Step 9 — Goroutine & Channel

| Concept | What it does |
|---------|-------------|
| `go func(){}()` | Starts a goroutine — lightweight concurrent thread (~2KB) |
| `ch := make(chan int)` | Creates a channel — pipe between goroutines |
| `ch <- val` / `val := <-ch` | Send / receive on a channel (blocks until other side is ready) |
| `make(chan int, N)` | Buffered channel — holds N values without blocking |
| `close(ch)` | Signals no more values will be sent |
| `for val := range ch` | Reads all values until the channel is closed |
| `select { case ... }` | Waits on multiple channels — first ready wins |
| `sync.WaitGroup` | `Add(1)`, `Done()`, `Wait()` — waits for N goroutines to finish |
| `sync.Mutex` | `Lock()`, `Unlock()` — prevents data races on shared state |
| `go test -race` | Race detector — finds concurrent access bugs |
| Worker pool pattern | N workers read from a shared jobs channel |
| `chan<- T` / `<-chan T` | Send-only / receive-only channel (directional) |

### Step 10 — Graceful Shutdown

| Concept | What it does |
|---------|-------------|
| `signal.Notify(ch, SIGINT, SIGTERM)` | Catches OS signals instead of dying |
| `SIGINT` / `SIGTERM` | Ctrl+C / `kill` command (also Kubernetes pod shutdown) |
| `http.Server{}` | Explicit server with timeouts (Read, Write, Idle) |
| `server.Shutdown(ctx)` | Stops accepting connections, waits for active requests |
| `context.WithTimeout` | Auto-cancels after a duration |
| `context.WithCancel` | Manually cancellable context |
| `ctx.Done()` channel | Closed when context is cancelled — use in `select` to react |
| `ctx.Err()` | `context.Canceled` or `context.DeadlineExceeded` |

### Step 11 — Docker Multi-Stage Build

| Concept | What it does |
|---------|-------------|
| Multi-stage build | Multiple `FROM` — build stage (big) → final stage (small) |
| `AS builder` | Names a stage for `COPY --from=builder` |
| Layer caching | `COPY go.mod` → `go mod download` → `COPY . .` — dependencies cached |
| `.dockerignore` | Excludes files from build context (like `.gitignore` for Docker) |
| `CGO_ENABLED=1` | Required for SQLite — needs C compiler (`gcc`) |
| `CMD ["./binary"]` | Exec form — process gets PID 1, receives signals correctly |
| Image size | ~1GB (with compiler) → ~30MB (binary only) |

### Step 12 — Generics

| Concept | What it does |
|---------|-------------|
| `[T any]` | Type parameter — a variable for types |
| `comparable` constraint | T must support `==` and `!=` |
| `any` constraint | No restriction — alias for `interface{}` |
| Custom constraint | `interface{ ~int \| ~float64 }` — union of types |
| `~int` | Accepts any type whose underlying type is int |
| Type inference | `Contains(nums, 3)` — Go infers `T=int` from arguments |
| Generic struct | `Pair[T, U]` — struct with type parameters |
| `Map[T, R]` | Transform `[]T` → `[]R` with a function |
| `Filter[T]` | Keep elements matching a predicate |

### Step 13 — Kubernetes Deployment

| Concept | What it does |
|---------|-------------|
| Deployment | Ensures N pod replicas are running; handles rolling updates |
| Service | Stable network endpoint for a set of pods |
| ConfigMap | Externalized config → injected as env vars |
| PVC | Persistent storage that survives pod restarts |
| `livenessProbe` | "Is it alive?" — fails → pod restarted |
| `readinessProbe` | "Can it serve traffic?" — fails → removed from Service |
| `startupProbe` | "Has it started?" — blocks liveness/readiness until done |
| `resources.requests` | Minimum guaranteed CPU/memory |
| `resources.limits` | Maximum allowed CPU/memory |
| `terminationGracePeriodSeconds` | Time between SIGTERM and SIGKILL |
| `os.Getenv` | Reads environment variables (from ConfigMap in K8s) |
| `t.Setenv` | Sets env var for one test only — auto-cleaned |

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/hello` | No | Health check |
| GET | `/slow` | No | 5s slow endpoint (shutdown testing) |
| GET | `/users` | Yes | List all users (optional `?city=` filter) |
| POST | `/users` | Yes | Create a user |
| GET | `/users/{id}` | Yes | Get user by ID |
| PUT | `/users/{id}` | Yes | Update user by ID |
| DELETE | `/users/{id}` | Yes | Delete user by ID |

**Auth header:** `X-API-Key: key-fatma-123`

## Running the Project

```bash
# Run locally
go run .

# Run tests
CGO_ENABLED=1 go test -v

# Run tests with race detector
CGO_ENABLED=1 go test -v -race

# Build Docker image
docker build -t go-training:latest .

# Run in Docker
docker run -p 8181:8181 -v go-data:/app/data go-training:latest

# Deploy to Kubernetes
kubectl apply -f deployment/

# Test the API
curl http://localhost:8181/hello
curl -H "X-API-Key: key-fatma-123" http://localhost:8181/users
curl -X POST -H "X-API-Key: key-fatma-123" -d '{"name":"Elif","age":22,"city":"Istanbul"}' http://localhost:8181/users
```

## Test Count: 59

```
apperror_test.go     — 5 tests  (custom error types, writeError, errors.As)
auth_test.go         — 6 tests  (no key, invalid key, valid key, public route, context)
config_test.go       — 4 tests  (defaults, env vars, getEnv)
concurrency_test.go  — 7 tests  (goroutine, channel, buffered, range, select, worker pool, mutex)
generics_test.go     — 16 tests (Contains, Map, Filter, Sum, Pair, Response, Keys)
handler_test.go      — 12 tests (CRUD operations, city filter, validation, 405)
middleware_test.go    — 4 tests  (logger passthrough, status capture, panic recovery)
shutdown_test.go     — 5 tests  (graceful shutdown, context cancel/timeout)
```
