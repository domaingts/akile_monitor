FROM golang:1.23-alpine3.20 AS builder

COPY . .

RUN CGO_ENABLE=0 go build -o /akile-monitor ./main.go

FROM alpine:3.20 AS dist

COPY --from=builder /akile-monitor /akile-monitor

ENTRYPOINT ["/akile-monitor"]
