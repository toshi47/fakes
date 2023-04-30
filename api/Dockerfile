FROM golang:1.19.5

RUN go version
ENV GOPATH=/

RUN apt-get update

WORKDIR /api

COPY ./ ./

RUN go mod download
RUN go build -o api ./main.go

CMD ["./api"]