package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/grpcpulse/internal/checker"
	"github.com/yourorg/grpcpulse/internal/config"
	"github.com/yourorg/grpcpulse/internal/metrics"
	"github.com/yourorg/grpcpulse/internal/scheduler"
	"github.com/yourorg/grpcpulse/internal/server"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	m := metrics.New()

	chk, err := checker.New(cfg)
	if err != nil {
		log.Fatalf("failed to create checker: %v", err)
	}

	sched := scheduler.New(scheduler.DefaultConfig(), chk, m)

	srv := server.New(server.DefaultConfig(), m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("metrics server error: %v", err)
		}
	}()

	go sched.Start(ctx)

	log.Printf("grpcpulse started — checking %s every %s", cfg.Address, cfg.Interval)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down grpcpulse...")
	cancel()
	sched.Stop()
	srv.Stop()
}
