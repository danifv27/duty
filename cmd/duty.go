package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danifv27/duty/internal/duty"

	log "github.com/gomicro/ledger"
)

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := duty.ServeMetrics(ctx); err != nil {
			log.Errorf("failed to serve metrics: +%v", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := duty.Serve(ctx); err != nil {
			log.Errorf("failed to serve: +%v", err)
		}
	}()
	wg.Wait()
	log.Info("Duty finished")
}
