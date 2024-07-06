FROM golang:latest

COPY ./ ./

RUN go mod download
RUN go build -o main main.go
CMD ["./main"]