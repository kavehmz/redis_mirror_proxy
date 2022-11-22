package main

import (
	"log"

	"github.com/tidwall/redcon"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

var mirrorDoQueue chan rdbDo

func main() {
	log.Printf("started server at %s", *proxyRedis)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println("Received signal", sig)
		// Reload config
		// config(...)
		done <- true
	}()

	// Create a channel do mirror commands with capacity of *buffer to receive the mirrored commands
	mirrorDoQueue = make(chan rdbDo, *buffer)
	// Starting a separate goroutine to process the mirror commands
	go mirrorDo()
	err := redcon.ListenAndServe(*proxyRedis,
		redisCommand,
		redisConnect,
		redisClose,
	)
	if err != nil {
		log.Fatal(err)
	}
}
