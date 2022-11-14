package main

import "flag"

var mainRedis *string
var mirrorRedis *string
var addr *string

func init() {
	mainRedis = flag.String("main", ":6379", "connection string for the main redis")
	mirrorRedis = flag.String("mirror", ":6381", "connection string for the main redis")
	addr = flag.String("addr", ":6380", "connection string for the mirroring service")
	flag.Parse()
}
