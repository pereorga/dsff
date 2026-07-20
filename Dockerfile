# --- Build Stage ---
FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

WORKDIR /app

COPY go/go.mod go/go.sum ./

RUN go mod download

COPY go/ .

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dsff

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
