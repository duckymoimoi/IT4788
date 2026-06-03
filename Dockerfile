# ==============================================================
# HOSPITAL NAVIGATION SYSTEM - Multi-stage Docker Build
# ==============================================================

# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cai dat dependencies truoc (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /hospital ./cmd/main.go

# ---- Stage 2: Build MAPF solver ----
FROM alpine:3.21 AS mapf-builder

RUN apk add --no-cache build-base cmake boost-dev boost-static spdlog-dev linux-headers

WORKDIR /src
COPY mapf/ ./mapf/
RUN cmake -B /src/mapf/build_docker /src/mapf -DPYTHON=false -DCMAKE_BUILD_TYPE=Release \
    && cmake --build /src/mapf/build_docker --target lifelong -j2

# ---- Stage 3: Runtime ----
FROM alpine:3.21

# Cai ca-certificates cho HTTPS (TTS Google Translate)
RUN apk --no-cache add ca-certificates tzdata libstdc++ boost1.84-libs spdlog

WORKDIR /app

# Copy binary tu builder
COPY --from=builder /hospital .
COPY --from=mapf-builder /src/mapf/build_docker/lifelong ./mapf_solver/lifelong

# Copy data files (map, output.json). /app/data may be shadowed by a
# persistent Docker volume, so keep a second immutable copy for boot repair.
COPY data/ ./data/
COPY data/ ./seed_data/

# Tao thu muc cho uploads va audio
RUN mkdir -p uploads audio

# Expose port
EXPOSE 8080

# Environment variables (override khi chay)
ENV APP_ENV=production
ENV PORT=8080
ENV DB_DSN="host=db user=postgres password=postgres dbname=hospital port=5432 sslmode=disable"
ENV MAPF_SOLVER_PATH="/app/mapf_solver/lifelong"

# Health check
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/sys/check_version?platform=docker\&version=1.0.0 || exit 1

# Run
CMD ["./hospital"]
