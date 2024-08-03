FROM golang:1.23-rc-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o apieng

FROM alpine

RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/apieng /apieng
COPY --from=builder /app/templates /templates
COPY --from=builder /app/metrics.db /metrics.db

EXPOSE 8080

CMD ["/apieng"]
