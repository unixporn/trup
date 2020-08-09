FROM golang:1.14-alpine
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go mod download && \
    go build -o main .
CMD ["/app/main"]
