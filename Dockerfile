# Build Server
FROM golang:latest as BUILD

WORKDIR /app

# Download
RUN mkdir speedtest
COPY speedtest/go.mod speedtest/go.mod
COPY speedtest/go.sum speedtest/go.sum

WORKDIR /app/speedtest
RUN go mod download

# Build
COPY speedtest /app/speedtest
WORKDIR /app/speedtest
RUN CGO_ENABLED=0 GOOS=linux go build -tags release -o /speedtest main.go

FROM thewh1teagle/lens
WORKDIR /app
COPY --from=BUILD /speedtest /speedtest


