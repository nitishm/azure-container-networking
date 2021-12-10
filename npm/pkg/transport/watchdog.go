package transport

import (
	"context"

	"google.golang.org/grpc/stats"
)

// Watchdog is a stats handler that watches for connection and RPC events.
// It implements the gRPC stats.Handler interface.
type Watchdog struct {
	// deregCh is used by the Watchdog to signal the Watchdog to deregister a remote address/client
	deregCh chan<- string
}

// NewWatchdog creates a new Watchdog instance
func NewWatchdog(deregCh chan<- string) stats.Handler {
	return &Watchdog{
		deregCh: deregCh,
	}
}

func (h *Watchdog) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *Watchdog) HandleRPC(ctx context.Context, _ stats.RPCStats) {
	_ = ctx
}

func (h *Watchdog) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	// Add the remote address to the context so that we can use it during a connection end event
	return context.WithValue(ctx, remoteAddrContextKey, info.RemoteAddr.String())
}

// HandleConn processes the Conn stats.
func (h *Watchdog) HandleConn(c context.Context, s stats.ConnStats) {
	switch s.(type) {
	// Watch for connection end events
	case *stats.ConnEnd:
		remoteAddr := c.Value(remoteAddrContextKey).(string)
		h.deregCh <- remoteAddr
	}
}
