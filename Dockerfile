# Container image that runs the code
FROM golang:1.21-alpine

WORKDIR /application

EXPOSE 9092

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./flux ./cmd/broker/*.go

RUN chmod +x ./flux

CMD ["./flux"]