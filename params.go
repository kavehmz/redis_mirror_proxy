package main

import "flag"

var mainRedis *string
var mirrorRedis *string
var proxyRedis *string
var buffer *int

func init() {
	// Load configuration from file
	mainRedis, mirrorRedis, proxyRedis = config()
	// Override undefined values per defaults
	if *mainRedis == ":" { *mainRedis = ":6379" }
	if *mirrorRedis == ":" { *mirrorRedis = ":6381" }
	if *proxyRedis == ":" { *proxyRedis = ":6380" }
	// Command line arguments has priority over configuration file
	mainRedis = flag.String("main", *mainRedis, "connection string for the main redis")
	mirrorRedis = flag.String("mirror", *mirrorRedis, "connection string for the main redis")
	proxyRedis = flag.String("addr", *proxyRedis, "connection string for the mirroring service")
	buffer = flag.Int("buff", 500, "Buffer size for mirror queue. If it is full Do commands will be ignored")

	flag.Parse()
}
