FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o safi-budget ./cmd/safi-budget

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/safi-budget ./safi-budget
COPY --from=builder /app/internal/data ./internal/data
COPY --from=builder /app/web ./web
ENV PORT=8080
EXPOSE 8080
CMD ["./safi-budget"]
