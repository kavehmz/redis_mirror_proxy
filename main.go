package main

import (
	"log"

	"github.com/tidwall/redcon"
)

var mirrorQueue chan mirrorCmd
var mainQueue chan mainCmd

func main() {
	log.Printf("Started server at %s", *addr)
	// Create a channel do mirror commands with capacity of *buffer to receive the mirrored commands
	mirrorQueue = make(chan mirrorCmd, *buffer)
	mainQueue = make(chan mainCmd, *buffer)

	initQueues()
	err := redcon.ListenAndServe(*addr,
		redisCommand,
		redisConnect,
		redisClose,
	)
	if err != nil {
		log.Fatal(err)
	}
}
