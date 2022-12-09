package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/redcon"
)

var syncMirror []chan bool
var contMirror []chan bool
var doneMirror []chan bool
var pauseMain []chan bool
var contMain []chan bool

type mirrorCmd struct {
	rdb     redis.Conn
	cmdArgs []interface{}
}

type mainCmd struct {
	conn    redcon.Conn
	cmdArgs []interface{}
}

var N = 10

func initQueues() {
	for i := 0; i < N; i++ {
		syncMirror = append(syncMirror, make(chan bool))
		contMirror = append(contMirror, make(chan bool))
		doneMirror = append(doneMirror, make(chan bool))
		pauseMain = append(pauseMain, make(chan bool))
		contMain = append(contMain, make(chan bool))
		go mainDo(i)
		go mirrorDo(i)
	}
}

// This function starts as a goroutine. A concurrent process.
// mirrorDo loops forever and run the queued commands against the mirror redis
func mirrorDo(i int) {
	for {
		select {
		case next := <-mirrorQueue:
			_, err := next.rdb.Do(next.cmdArgs[0].(string), next.cmdArgs[1:]...)
			if err != nil {
				log.Println("Mirror command failed:", err.Error())
			}
		case <-syncMirror[i]:
			doneMirror[i] <- true
			<-contMirror[i]
		}
	}
}

func mainDo(i int) {
	for {
		select {
		case next := <-mainQueue:
			res, err := pickMain(next.conn).Do(next.cmdArgs[0].(string), next.cmdArgs[1:]...)
			if err == nil {
				// This will queue the the mirror commands until the channel reaches its capacity.
				// If channel is full, commands will be ignored and we get log and error
				select {
				case mirrorQueue <- mirrorCmd{rdb: pickMirror(next.conn), cmdArgs: next.cmdArgs}:
				default:
					log.Println("Queue full, skipping", next.cmdArgs)
				}
			}

			next.conn.Context().(RedisSettings).replyChan <- reply{res: res, err: err}
		case <-pauseMain[i]:
			<-contMain[i]
		}

	}
}
