# --- Build Stage ---
FROM golang:1.25-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 AS builder

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

EXPOSE 80

CMD ["./dsff"]
