FROM golang:1.14-alpine

RUN mkdir /app

WORKDIR /app

ADD go.mod go.sum /app/
RUN go mod download

COPY db /app/db
COPY command /app/command
COPY ctx /app/ctx
COPY misc /app/misc
COPY eventhandler /app/eventhandler
COPY routine /app/routine
COPY *.go /app/

RUN go build -o main .

CMD ["/app/main"]
