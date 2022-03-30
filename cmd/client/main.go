package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/idekpas/grpc-task/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr    = flag.String("addr", "server:9081", "the address to connect to")
	timeout = flag.Int("timeout", 60, "Client context timeout in seconds")
)

func getContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	type ctxKey string
	k := ctxKey("ID")
	uuid, err := uuid.NewRandom()
	if err != nil {
		return ctx, cancel
	}
	ctx = context.WithValue(ctx, k, uuid)
	return ctx, cancel
}

func printName(c pb.NameServiceClient, ctx context.Context) {
	gnr, err := c.GetName(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get name: %v", err)
	}
	log.Printf("Server name: %s", gnr.GetName())
}

func sendName(c pb.NameServiceClient, ctx context.Context) {
	name := "Name" + strconv.Itoa(rand.Int())
	_, err := c.SetName(ctx, &pb.NameRequest{Name: name})
	if err != nil {
		log.Fatalf("could not set name: %v", err)
	}
	log.Printf("Set name: %s", name)
}

func printNames(c pb.NameServiceClient, ctx context.Context) {
	log.Printf("Getting names:")
	stream, err := c.GetNameStream(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("%v.GetNameStream(_) = _, %v", c, err)
	}

	for {
		gnr, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.GetNameStream(_) = _, %v", c, err)
		}
		log.Printf("Name: %q", gnr.GetName())
	}
}

func runSendNames(c pb.NameServiceClient, namesCount int, ctx context.Context) {
	log.Printf("Sending names")

	var names []*pb.NameRequest
	for i := 0; i < namesCount; i++ {
		names = append(names, &pb.NameRequest{Name: "Name" + strconv.Itoa(i)})
	}

	stream, err := c.SetNameStream(ctx)
	if err != nil {
		log.Fatalf("%v.SetNameStream(_) = _, %v", c, err)
	}
	for _, name := range names {
		if err := stream.Send(name); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, name, err)
		}
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Sending complete")
}

func main() {
	flag.Parse()
	ctx, cancel := getContext()
	defer cancel()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNameServiceClient(conn)

	//sendName(c)
	//sendName(c)
	//printName(c)

	var wg = &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		printNames(c, ctx)
	}()
	go func() {
		defer wg.Done()
		runSendNames(c, 90, ctx)
	}()
	wg.Wait()
}
