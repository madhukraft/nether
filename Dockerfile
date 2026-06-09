FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/madhukraft/nether/internal/cmd.Version=$VERSION" \
    -o /nether ./cmd/nether

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /nether /usr/local/bin/nether
WORKDIR /data
ENTRYPOINT ["nether"]
