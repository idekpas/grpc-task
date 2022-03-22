package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/idekpas/grpc-task/nameservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr    = flag.String("addr", "localhost:9081", "the address to connect to")
	timeout = flag.Int("timeout", 60, "Client timeout in seconds")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNameServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	gnr, err := c.GetName(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get name: %v", err)
	}
	log.Printf("Get initial server name: %s", gnr.GetName())

	for {
		name := "New Name " + strconv.Itoa(rand.Int())
		_, err = c.SetName(ctx, &pb.NameRequest{Name: name})
		if err != nil {
			log.Fatalf("could not set name: %v", err)
		}
		log.Printf("Set name: %s", name)

		gnr, err = c.GetName(ctx, &pb.Empty{})
		if err != nil {
			log.Fatalf("could not get name: %v", err)
		}
		log.Printf("Get name: %s", gnr.GetName())

		time.Sleep(3 * time.Second)
	}
}
