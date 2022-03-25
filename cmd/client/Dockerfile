FROM golang:1.18

RUN mkdir -p /app/client

ADD . /app/client/cmd
ADD ./pb/. /app/client/pb

WORKDIR /app/client/cmd

RUN go run cmd/client/main.go

CMD ["/app/client/cmd/main"]