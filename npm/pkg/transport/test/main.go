package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azure-container-networking/npm/pkg/transport"
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

	m := transport.NewManager(8080)

	m.Start(ctx)
}
