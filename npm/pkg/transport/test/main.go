package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azure-container-networking/npm/pkg/transport"
)

const (
	port = 8080
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	m := transport.NewManager(port)

	if err := m.Start(ctx); err != nil {
		panic(err)
	}
}
