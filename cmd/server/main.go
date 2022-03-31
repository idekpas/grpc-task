package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/idekpas/grpc-task/cmd/server/messaging"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	pb "github.com/idekpas/grpc-task/pb"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9081, "The server port")
)

type Messager interface {
	Subscribe(id string) chan string
	Publish(ch chan string, names ...string)
	Pull(ch chan string) string
}

type nameServiceServer struct {
	pb.UnimplementedNameServiceServer

	msg  Messager
	mu   sync.Mutex
	name string
}

func NewNameServiceServer(msg Messager) *nameServiceServer {
	return &nameServiceServer{msg: msg}
}

func getID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Fatalf("No request id received")
		return uuid.Nil.String()
	}

	if v := md.Get("ID"); v != nil {
		id, err := uuid.Parse(v[0])
		if err != nil {
			log.Println("id is null")
			return uuid.Nil.String()
		}
		return id.String()
	}
	log.Println("no id")
	return uuid.Nil.String()
}

func (s *nameServiceServer) GetName(_ context.Context, _ *pb.Empty) (*pb.NameResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return &pb.NameResponse{Name: s.name}, nil
}

func (s *nameServiceServer) SetName(_ context.Context, req *pb.NameRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.name = req.Name
	return &pb.Empty{}, nil
}

func (s *nameServiceServer) GetNameStream(_ *pb.Empty, stream pb.NameService_GetNameStreamServer) error {
	ctx := stream.Context()
	dataCh := s.msg.Subscribe(getID(ctx))
	for name := range dataCh {
		if err := stream.Send(&pb.NameResponse{Name: name}); err != nil {
			return err
		}

		err := ctx.Err()
		if err != nil {
			return nil //connection closed
		}
	}

	return nil
}

func (s *nameServiceServer) SetNameStream(stream pb.NameService_SetNameStreamServer) error {
	ctx := stream.Context()
	dataCh := s.msg.Subscribe(getID(ctx))
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Empty{})
		}
		if err != nil {
			return err
		}

		s.msg.Publish(dataCh, r.Name)
	}
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNameServiceServer(s, NewNameServiceServer(messaging.NewService()))
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
