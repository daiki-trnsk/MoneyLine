FROM golang:1.25.3-alpine

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8000

CMD ["./main"]