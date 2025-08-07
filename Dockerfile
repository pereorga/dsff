# --- Build Stage ---
FROM golang:1.24-alpine@sha256:daae04ebad0c21149979cd8e9db38f565ecefd8547cf4a591240dc1972cf1399 AS builder

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
