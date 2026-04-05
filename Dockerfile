FROM golang:1.25-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Version is passed as a build arg (defaults to "dev").
ARG VERSION=dev

# Build both binaries as static executables (pure Go, no CGO needed).
# Bookworm base matches TrueNAS SCALE's Debian 12 for test environment.
# GOAMD64=v1 ensures baseline x86-64 compatibility.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
    go build -ldflags="-s -w -X main.version=${VERSION}" -o /out/qnaplcd ./cmd/qnaplcd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
    go build -ldflags="-s -w" -o /out/qnaplcd-test ./cmd/qnaplcd-test

# Run tests.
RUN go test ./...

# Final stage — just the binaries (no OS needed for static builds).
FROM scratch
COPY --from=builder /out/qnaplcd /qnaplcd
COPY --from=builder /out/qnaplcd-test /qnaplcd-test
CMD ["/qnaplcd"]
