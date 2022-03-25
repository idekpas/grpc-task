FROM golang:1.18

RUN mkdir -p /app/server

ADD . /app/server/cmd
ADD ./pb/. /app/server/pb

WORKDIR /app/server/cmd

RUN go run cmd/server/main.go

CMD ["/app/server/cmd/main"]
EXPOSE 9082