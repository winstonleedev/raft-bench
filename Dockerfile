FROM golang:alpine AS builder
RUN mkdir /build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kv-bench .

FROM golang:bullseye as build-env
WORKDIR /app
COPY --from=builder /build/kv-bench .

CMD ["bash"]

