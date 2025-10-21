package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/PriestYKing/blitzcache"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	addr := flag.String("addr", ":6380", "Server address")
	shards := flag.Int("shards", 256, "Number of shards")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("BlitzCache %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	log.Printf("Starting BlitzCache %s", version)

	cache := blitzcache.NewCache(*shards)
	server := blitzcache.NewServer(*addr, cache)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cache.Close()
		os.Exit(0)
	}()

	log.Fatal(server.Start())
}
