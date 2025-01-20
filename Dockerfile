FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

COPY .env .

EXPOSE 8080

CMD ["./main"]