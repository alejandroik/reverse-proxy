FROM golang:1.17.0

WORKDIR /app

RUN export GO111MODULE=on

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o rev-proxy .

EXPOSE 8080

CMD ["./rev-proxy"]