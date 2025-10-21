package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/PriestYKing/blitzcache"
)

func main() {
	addr := flag.String("addr", ":6380", "Server address")
	shards := flag.Int("shards", 256, "Number of shards")
	flag.Parse()

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
