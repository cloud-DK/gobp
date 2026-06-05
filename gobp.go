package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/cloud-DK/gobp/tui"
)

func main() {
	startTime := time.Now()
	// Startup global context.
	ctx := context.Background()
	go shutdown(startTime, ctx) // Set up graceful shutdown
	defer func() {
		elapsed := time.Since(startTime)
		fmt.Printf("Execution took %s\n", elapsed)
	}()

	if err := tui.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func shutdown(startTime time.Time, ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		sig := <-s
		elapsed := time.Since(startTime)
		fmt.Printf("Shut down down signal received <%v>. \n ", sig)
		fmt.Printf("Shutting down after %s\n", elapsed)
		// clean up here
		ctx.Done()
		os.Exit(0)
	}()
}
