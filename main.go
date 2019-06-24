package main

import (
	"context"
	"log"
	// "os"
	"time"

	pb "./proto"
	"google.golang.org/grpc"
)

const (
	address = "ac.testnet.libra.org:8000"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewAdmissionControlClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 100)
	defer cancel()

	r, err := c.UpdateToLatestLedger(
		ctx,
		&pb.UpdateToLatestLedgerRequest{
			ClientKnownVersion: 0,
			RequestedItems: []*pb.RequestItem{
				&pb.RequestItem {
					RequestedItems: &pb.RequestItem_GetTransactionsRequest{
						&pb.GetTransactionsRequest{
							StartVersion: 0,
							Limit:        100,
							FetchEvents:  true,
						},
					},
				},
			},
		})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.String())
}
