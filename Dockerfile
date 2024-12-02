FROM golang:1.22-alpine AS builder

RUN apk add --no-cache make git

WORKDIR /app
COPY . .

ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o portscan-server cmd/server/main.go

FROM alpine:latest

RUN apk add --no-cache masscan

COPY --from=builder /app/portscan-server /app/portscan-server

COPY ./app.env /app/app.env

COPY ./wait-for.sh /app/wait-for.sh

WORKDIR /app

RUN chmod +x /app/wait-for.sh

ENTRYPOINT ["./portscan-server"]