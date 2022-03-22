package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/idekpas/grpc-task/nameservice"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9081, "The server port")
	name = flag.String("name", "Default", "Default name returned by server")
)

type nameServiceServer struct {
	pb.UnimplementedNameServiceServer

	mu   sync.Mutex
	name string
}

func NewNameServiceServer() *nameServiceServer {
	return &nameServiceServer{name: *name}
}

func (s *nameServiceServer) GetName(c context.Context, empty *pb.Empty) (*pb.NameResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return &pb.NameResponse{Name: s.name}, nil
}

func (s *nameServiceServer) SetName(c context.Context, request *pb.NameRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.name = request.Name

	return &pb.Empty{}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNameServiceServer(s, NewNameServiceServer())
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
