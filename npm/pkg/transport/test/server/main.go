package main

import (
	"flag"
	"fmt"
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

var port int

func main() {
	flag.IntVar(&port, "port", 8080, "Server listening port")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		<-sigs
		done <- struct{}{}
	}()

	go startServer(done, port)

	<-done
}

type DataplaneEventsServer struct {
	transportpb.UnimplementedDataplaneEventsServer
}

func (DataplaneEventsServer) Connect(m *transportpb.DatapathPodMetadata, stream transportpb.DataplaneEvents_ConnectServer) error {
	fmt.Printf("Received a Connect request from Client ID [%s]\n", m.GetId())
	fmt.Println("Sending stream of events")

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

func startServer(done chan struct{}, port int) {
	fmt.Println("Starting server")
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Printf("failed to listen on port %dwith err : %v", port, err)
		done <- struct{}{}
	}

	fmt.Printf("Listening on port %d\n", port)

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
		fmt.Printf("failed to listen on port 8080 with err : %v", err)
		done <- struct{}{}
	}
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
