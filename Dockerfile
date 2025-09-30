# --- Build Stage ---
FROM golang:1.25-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS builder

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
