package main

import (
	"flag"
)

var redisAddr1 *string
var redisAddr2 *string
var addr *string
var buffer *int

func init() {
	redisAddr1 = flag.String("redis1", ":6379", "connection string for the redis 1")
	redisAddr2 = flag.String("redis2", ":6381", "connection string for the redis 2")
	addr = flag.String("addr", ":6380", "connection string for the mirroring service")
	buffer = flag.Int("buff", 500, "Buffer size for mirror queue. If it is full Do commands will be ignored")
	flag.Parse()
}
