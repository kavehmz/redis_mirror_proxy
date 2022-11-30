package main

import "flag"

var mainRedis *string
var mirrorRedis *string
var addr *string
var buffer *int

func init() {
	mainRedis = flag.String("main", ":6379", "connection string for the main redis")
	mirrorRedis = flag.String("mirror", ":6381", "connection string for the mirror redis")
	addr = flag.String("addr", ":6380", "connection string for the mirroring service")
	buffer = flag.Int("buff", 500, "Buffer size for mirror queue. If it is full Do commands will be ignored")
	flag.Parse()
}
