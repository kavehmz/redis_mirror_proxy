package main

import "flag"

var connection1 *string
var connection2 *string
var mode *int
var addr *string
var buffer *int

func init() {
	connection1 = flag.String("connection1", ":6379", "connection string for the redis 1")
	connection2 = flag.String("connection2", ":6381", "connection string for the redis 2")
	mode = flag.Int("mode", 1, "main connection to use (1 or 2)")
	addr = flag.String("addr", ":6380", "connection string for the mirroring service")
	buffer = flag.Int("buff", 500, "Buffer size for mirror queue. If it is full Do commands will be ignored")
	flag.Parse()
}
