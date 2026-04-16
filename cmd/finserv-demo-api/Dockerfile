FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init finserv-demo-api && \
    go build -o finserv-demo-api .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /build/finserv-demo-api .
ENV PORT=9090
EXPOSE 9090
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:9090/health || exit 1
CMD ["./finserv-demo-api"]
