package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/darkit/process"
	"github.com/darkit/slog"
)

func main() {
	var (
		configFile string
	)

	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file path")
	flag.Parse()

	manager := process.NewManager()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := <-sigChan
		slog.Infof("Received signal: %v", sig)
		cancel()
	}()

	// Example process creation
	proc, err := manager.NewProcessCmd("echo 'Hello World'", nil)
	if err != nil {
		slog.Fatalf("Failed to create process: %v", err)
	}

	proc.Start(true)

	<-ctx.Done()
	manager.StopAllProcesses()
}
