FROM golang:1.14-alpine

RUN mkdir /app

WORKDIR /app

ADD go.mod go.sum /app/
RUN go mod download

ADD . /app
RUN go build -o main .

CMD ["/app/main"]
