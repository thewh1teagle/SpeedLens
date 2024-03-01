# Build Server
FROM golang:latest as BUILD

WORKDIR /app

COPY speedtest speedtest
WORKDIR /app/speedtest
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -tags release -o /speedtest main.go

FROM thewh1teagle/lens
WORKDIR /app
COPY --from=BUILD /speedtest /speedtest


