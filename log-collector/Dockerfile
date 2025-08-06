FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN go build -o log-collector

CMD ["./log-collector"]
