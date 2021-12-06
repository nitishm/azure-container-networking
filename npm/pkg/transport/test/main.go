package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	transportpb "github.com/Azure/azure-container-networking/npm/pkg/transport/pb"
	"github.com/fatih/structs"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	port = 8080
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{}, 1)

	go func() {
		s := <-sigs
		log.Printf("received signal: %s", s)
		done <- struct{}{}
	}()

	go func() {
		// We do not have a signal handler here
		// TODO: Break this out into two programs rather than having both run in the same thread/process
		if err := startServer(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
		done <- struct{}{}
	}()

	// Client code
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cc, err := grpc.DialContext(ctx, "localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	client := transportpb.NewDataplaneEventsClient(cc)
	clientID := "test-client"
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
			if err != nil || err == io.EOF {
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

type DataplaneEventsServer struct {
	transportpb.UnimplementedDataplaneEventsServer
}

func (DataplaneEventsServer) Connect(m *transportpb.DatapathPodMetadata, stream transportpb.DataplaneEvents_ConnectServer) error {
	for {
		err := stream.SendMsg(&transportpb.Events{
			Type:   transportpb.Events_APPLY,
			Object: transportpb.Events_IPSET,
			Event: []*transportpb.Event{{
				Data: testData(),
			}},
		})
		if err != nil {
			return fmt.Errorf("failed to send: %w", err)
		}
		time.Sleep(time.Second * 5)
	}
}

func startServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	// Start gRPC Server in background
	var opts []grpc.ServerOption = []grpc.ServerOption{
		grpc.MaxConcurrentStreams(100),
	}

	server := grpc.NewServer(opts...)

	transportpb.RegisterDataplaneEventsServer(
		server,
		// this is our own implementation of DataplaneEventsServer
		DataplaneEventsServer{},
	)

	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func testData() []*structpb.Struct {
	var output []*structpb.Struct
	for i := 0; i < 10; i++ {
		v := struct {
			Type    string
			Payload string
		}{
			Type:    fmt.Sprintf("IPSET-%d", i),
			Payload: fmt.Sprintf("172.17.0.%d/%d", i, rand.Uint32()%32),
		}

		m := structs.Map(v)

		data, _ := structpb.NewStruct(m)
		output = append(output, data)
	}
	return output
}
