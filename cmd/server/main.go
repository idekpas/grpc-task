package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/idekpas/grpc-task/cmd/server/messaging"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
	pb "github.com/idekpas/grpc-task/pb"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9081, "The server port")
	name = flag.String("name", "Default", "Default name returned by server")
)

type Messager interface {
	Subscribe(id string) chan string
	Publish(ch chan string, names ...string)
	Pull(ch chan string) string
}

type nameServiceServer struct {
	pb.UnimplementedNameServiceServer

	msg Messager
}

func NewNameServiceServer(msg Messager) *nameServiceServer {
	return &nameServiceServer{msg: msg}
}

func getID(ctx context.Context) string {
	type ctxKey string
	k := ctxKey("ID")
	if v := ctx.Value(k); v != nil {
		u, err := uuid.Parse(v.(string))
		if err != nil {
			return uuid.Nil.String()
		}
		return u.String()
	}
	return uuid.Nil.String()
}

func (s *nameServiceServer) GetName(ctx context.Context, _ *pb.Empty) (*pb.NameResponse, error) {
	dataCh := s.msg.Subscribe(getID(ctx))
	name := s.msg.Pull(dataCh)
	return &pb.NameResponse{Name: name}, nil
}

func (s *nameServiceServer) SetName(ctx context.Context, req *pb.NameRequest) (*pb.Empty, error) {
	dataCh := s.msg.Subscribe(getID(ctx))
	s.msg.Publish(dataCh, req.Name)
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
