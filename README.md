# grpc-task

go run client/main.go
go run server/main.go

protoc --go_out=nameservice --go_opt=paths=source_relative --go-grpc_out=nameservice --go-grpc_opt=paths=source_relative     service.proto



1. 