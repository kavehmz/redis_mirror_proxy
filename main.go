package main

import (
	"log"

	"github.com/tidwall/redcon"
)

func main() {
	log.Printf("started server at %s", *addr)
	err := redcon.ListenAndServe(*addr,
		redisCommand,
		redisConnect,
		redisClose,
	)
	if err != nil {
		log.Fatal(err)
	}
}
