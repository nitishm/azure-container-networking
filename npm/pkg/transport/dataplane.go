package transport

import (
	"context"
	"fmt"

	"github.com/Azure/azure-container-networking/npm/pkg/protos"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// DataplaneEventsClient is a client for the DataplaneEvents service
type DataplaneEventsClient struct {
	protos.DataplaneEventsClient
	pod        string
	node       string
	serverAddr string

	outCh chan *protos.Events
}

func NewDataplaneEventsClient(ctx context.Context, pod, node, addr string) (*DataplaneEventsClient, error) {
	// TODO Make this secure
	cc, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return &DataplaneEventsClient{
		DataplaneEventsClient: protos.NewDataplaneEventsClient(cc),
		pod:                   pod,
		node:                  node,
		serverAddr:            addr,
		outCh:                 make(chan *protos.Events),
	}, nil
}

func (c *DataplaneEventsClient) EventsChannel() <-chan *protos.Events {
	return c.outCh
}

func (c *DataplaneEventsClient) Start(ctx context.Context, stopCh <-chan struct{}) error {
	clientMetadata := &protos.DatapathPodMetadata{
		PodName:  c.pod,
		NodeName: c.node,
	}

	opts := []grpc.CallOption{}
	connectClient, err := c.Connect(ctx, clientMetadata, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to dataplane events server: %w", err)
	}

	return c.run(ctx, connectClient, stopCh)
}

func (c *DataplaneEventsClient) run(ctx context.Context, connectClient protos.DataplaneEvents_ConnectClient, stopCh <-chan struct{}) error {
	for {
		select {
		case <-ctx.Done():
			klog.Errorf("context done: %v", ctx.Err())
			return ctx.Err()
		case <-stopCh:
			klog.Info("Received message on stop channel. Stopping transport client")
			return nil
		default:
			event, err := connectClient.Recv()
			if err != nil {
				klog.Errorf("failed to receive event: %v", err)
				return err
			}

			c.outCh <- event
		}
	}
}
