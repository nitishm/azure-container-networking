package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	transportpb "github.com/Azure/azure-container-networking/npm/pkg/transport/pb"
	"google.golang.org/grpc"
)

var (
	clientID string
	addr     string
)

func main() {
	flag.StringVar(&clientID, "client-id", "", "Client identifier")
	flag.StringVar(&addr, "server-addr", "localhost:8080", "Server address")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{}, 1)

	go func() {
		s := <-sigs
		log.Printf("received signal: %s", s)
		done <- struct{}{}
	}()

	// Client code
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cc, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	client := transportpb.NewDataplaneEventsClient(cc)
	stream, err := client.Connect(
		context.Background(),
		&transportpb.DatapathPodMetadata{Id: clientID})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	for {
		select {
		case <-done:
			cancel()
			os.Exit(1)
		default:
			ev, err := stream.Recv()
			if err != nil || errors.Is(err, io.EOF) {
				log.Fatalf("failed to receive: %v", err)
			}

			fmt.Printf(
				"[Client ID: %s] Received event type %s object type %s: \n",
				clientID,
				ev.GetType(),
				ev.GetObject(),
			)

			for _, e := range ev.GetEvent() {
				for _, d := range e.GetData() {
					eventAsMap := d.AsMap()
					fmt.Printf("%s: %s\n", eventAsMap["Type"], eventAsMap["Payload"])
				}
			}
		}
	}
}
