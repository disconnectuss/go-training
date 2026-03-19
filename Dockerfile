# Step 11: Multi-Stage Build — Production-ready Docker image
#
# MULTI-STAGE BUILD: Multiple FROM statements in one Dockerfile
# Stage 1 (builder): Has all build tools (Go compiler, gcc, etc.) — BIG image
# Stage 2 (final):   Only has the compiled binary — SMALL image
#
# Why? Go compiles to a single binary. We don't need the compiler in production.
# Result: Image goes from ~1GB to ~30MB

# ============================================
# STAGE 1: Build the application
# ============================================
FROM golang:1.26-alpine AS builder
# golang:1.26-alpine = Go compiler on Alpine Linux (small base image)
# "AS builder" names this stage — we reference it later with COPY --from=builder

WORKDIR /app

# CGO_ENABLED=1 because SQLite driver (go-sqlite3) needs C compiler
# go-sqlite3 is a CGO package — it wraps C code, so we need gcc
ENV CGO_ENABLED=1

# Install C compiler and SQLite headers (needed for go-sqlite3)
# apk = Alpine's package manager (like apt for Ubuntu)
RUN apk add --no-cache gcc musl-dev

# COPY go.mod and go.sum FIRST, then download dependencies
# Docker caches each layer. If go.mod doesn't change, dependencies are cached!
# This means "go mod download" is skipped on rebuilds if only .go files changed
# Without this trick, every code change re-downloads ALL dependencies
COPY go.mod go.sum ./
RUN go mod download

# NOW copy the source code
# This layer changes every time we edit code, but the dependency layer above is cached
COPY . .

# Build the binary
# -o go-training = output file name
# CGO builds are OS-specific, so we don't set GOOS/GOARCH — use the container's arch
RUN go build -o go-training .

# ============================================
# STAGE 2: Create the minimal production image
# ============================================
FROM alpine:3.21
# Fresh Alpine image — NO Go compiler, NO source code, NO build tools

# SQLite needs C library at runtime — Alpine's musl is already included
# But we add ca-certificates for HTTPS calls (if needed in future)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# COPY --from=builder: Take ONLY the compiled binary from stage 1
# Everything else (source code, Go compiler, gcc) is thrown away!
COPY --from=builder /app/go-training .

# Create a directory for the SQLite database file
# In production, you'd mount a volume here: docker run -v db-data:/app/data
RUN mkdir -p /app/data

# Document which port the app listens on
# EXPOSE doesn't publish the port — it's documentation for humans and tools
EXPOSE 8181

# Run the binary
# Use array form (exec form) — the process gets PID 1 and receives signals correctly
# This is critical for graceful shutdown! Without exec form, signals go to shell, not our app
CMD ["./go-training"]
