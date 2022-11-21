package main

import (
	"log"

	"github.com/tidwall/redcon"
)

var mirrorDoQueue chan rdbDo

func main() {
	log.Printf("started server at %s", *addr)
	// Create a channel do mirror commands with capacity of *buffer to receive the mirrored commands
	mirrorDoQueue = make(chan rdbDo, *buffer)
	// Starting a separate goroutine to process the mirror commands
	go mirrorDo()
	err := redcon.ListenAndServe(*addr,
		redisCommand,
		redisConnect,
		redisClose,
	)
	if err != nil {
		log.Fatal(err)
	}
}
