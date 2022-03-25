package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/idekpas/grpc-task/pb"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"sync"
)

var (
	port = flag.Int("port", 9081, "The server port")
	name = flag.String("name", "Default", "Default name returned by server")
)

type nameServiceServer struct {
	pb.UnimplementedNameServiceServer

	mu    sync.Mutex
	queue []string
}

func NewNameServiceServer() *nameServiceServer {
	queue := make([]string, 0)
	queue = append(queue, *name)
	return &nameServiceServer{queue: queue}
}

func (s *nameServiceServer) GetName(c context.Context, empty *pb.Empty) (*pb.NameResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := s.queue[0]
	s.queue = s.queue[1:]
	return &pb.NameResponse{Name: name}, nil
}

func (s *nameServiceServer) SetName(c context.Context, request *pb.NameRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queue = append(s.queue, request.Name)
	return &pb.Empty{}, nil
}

func (s *nameServiceServer) GetNameStream(empty *pb.Empty, stream pb.NameService_GetNameStreamServer) error {
	for _, name := range s.queue {
		s.queue = s.queue[1:]
		if err := stream.Send(&pb.NameResponse{Name: name}); err != nil {
			return err
		}
	}
	return nil
}

func (s *nameServiceServer) SetNameStream(stream pb.NameService_SetNameStreamServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Empty{})
		}
		if err != nil {
			return err
		}

		s.queue = append(s.queue, r.Name)
	}
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
