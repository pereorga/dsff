# --- Build Stage ---
FROM golang:1.24-alpine@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS builder

WORKDIR /app

COPY go/go.mod go/go.sum ./

RUN go mod download

COPY go/ .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dsff

# --- Final Image ---
FROM scratch

WORKDIR /app

# include built binary
COPY --from=builder /app/dsff .

# include JSON data file
COPY data.json.gz .

# include static assets
COPY public/ public/

ARG PORT=80
ENV PORT=${PORT}
EXPOSE ${PORT}

CMD ["./dsff"]
