FROM golang:1.21.4

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install github.com/codegangsta/gin@latest

EXPOSE 8080

CMD ["gin", "--appPort", "8080", "run", "main.go"]
