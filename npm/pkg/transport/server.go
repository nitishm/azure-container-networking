package transport

import (
	"time"

	"github.com/Azure/azure-container-networking/npm/pkg/transport/pb"
	"google.golang.org/grpc/peer"
)

// clientStreamConnection represents a client stream connection
type clientStreamConnection struct {
	stream pb.DataplaneEvents_ConnectServer
	*pb.DatapathPodMetadata
	addr      string
	timestamp int64
}

// Addr returns the address of the client
func (c clientStreamConnection) String() string {
	return c.addr
}

// DataplaneEventsServer is the gRPC server for the DataplaneEvents service
type DataplaneEventsServer struct {
	pb.UnimplementedDataplaneEventsServer
	regCh chan<- clientStreamConnection
}

// NewServer creates a new DataplaneEventsServer instance
func NewServer(ch chan clientStreamConnection) *DataplaneEventsServer {
	return &DataplaneEventsServer{
		regCh: ch,
	}
}

// Connect is called when a client connects to the server
func (d *DataplaneEventsServer) Connect(m *pb.DatapathPodMetadata, stream pb.DataplaneEvents_ConnectServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return ErrNoPeer
	}

	conn := clientStreamConnection{
		DatapathPodMetadata: m,
		stream:              stream,
		addr:                p.Addr.String(),
		timestamp:           time.Now().Unix(),
	}

	// Add stream to the list of active streams
	d.regCh <- conn
	return nil
}
