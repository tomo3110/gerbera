FROM golang:1.22-bookworm AS builder

WORKDIR /src

# Copy the entire gerbera repo (needed for replace directive)
COPY go.mod go.sum ./
COPY . .

WORKDIR /src/example/sns
RUN go mod download
RUN CGO_ENABLED=0 go build -o /sns .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /sns /usr/local/bin/sns

EXPOSE 8930

CMD ["sns"]
