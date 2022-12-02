package main

import (
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/redcon"
)

type RedisSettings struct {
	conns map[int]redis.Conn
}

type rdbDo struct {
	rdb     redis.Conn
	cmdArgs []interface{}
}

var ps redcon.PubSub

// This function will process the incoming messages from client. It will act as a multiplexer.
// To get more information refer to:
// https://redis.io/docs/reference/protocol-spec
// https://pkg.go.dev/github.com/gomodule/redigo/redis#hdr-Executing_Commands
func redisCommand(conn redcon.Conn, cmd redcon.Command) {
	var rdbMain, rdbMirror redis.Conn
	if *mode == 1 {
		rdbMain = conn.Context().(RedisSettings).conns[1]
		rdbMirror = conn.Context().(RedisSettings).conns[2]
        } else {
		rdbMain = conn.Context().(RedisSettings).conns[2]
		rdbMirror = conn.Context().(RedisSettings).conns[1]
	}

	cmdArgs := []interface{}{}
	for _, v := range cmd.Args {
		cmdArgs = append(cmdArgs, string(v))
	}
	res, err := rdbMain.Do(cmdArgs[0].(string), cmdArgs[1:]...)
	if err != nil {
		conn.WriteError(err.Error())
		return
	}

	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		// This will queue the the mirror commands until the channle reaches its capacity.
		// If channel is full, commands will be ignored
		select {
		case mirrorDoQueue <- rdbDo{rdb: rdbMirror, cmdArgs: cmdArgs}:
		default:
			log.Println("Queue full, skipping", cmdArgs)
		}

		if res == nil {
			conn.WriteNull()
			return
		}
		printValue(res, conn)
	case "subscribe", "psubscribe":
		err := rdbMain.Send(cmdArgs[0].(string), cmdArgs[1:]...)
		if err != nil {
			log.Println("Errors", err)
			conn.WriteError(err.Error())
			return
		}
		rdbMain.Flush()

		command := strings.ToLower(string(cmd.Args[0]))
		for i := 1; i < len(cmd.Args); i++ {
			if command == "psubscribe" {
				ps.Psubscribe(conn, string(cmd.Args[i]))
			} else {
				ps.Subscribe(conn, string(cmd.Args[i]))
			}
		}

		for {
			res, err = rdbMain.Receive()
			if err != nil {
				log.Println("Errors", err)
				conn.WriteError(err.Error())
				return
			}
			cmd := string(res.([]interface{})[0].([]byte))
			log.Printf("received %+v, %s\n", res, cmd)
			if cmd == "message" {
				ps.Publish(string(res.([]interface{})[1].([]byte)), string(res.([]interface{})[2].([]byte)))
			}
			if cmd == "pmessage" {
				ps.Publish(string(res.([]interface{})[2].([]byte)), string(res.([]interface{})[3].([]byte)))
			}
		}
	}
}

func printValue(res interface{}, conn redcon.Conn) {
	switch v := res.(type) {
	case int64:
		conn.WriteInt64(v)
	case string:
		conn.WriteString(v)
	case []byte:
		conn.WriteBulk(v)
	case []interface{}:
		conn.WriteArray(len(v))
		for _, val := range v {
			switch v := val.(type) {
			case int64:
				conn.WriteInt64(v)
			case string:
				conn.WriteString(v)
			case []byte:
				conn.WriteBulk(v)
			case []interface{}:
				printValue(v, conn)
			default:
				log.Printf("This is an unknow type! %T", v)
			}
		}
	default:
		log.Printf("This is an unknow type! %T", v)
	}
}

func redisConnect(conn redcon.Conn) bool {
	c1, err := redis.Dial("tcp", *connection1)
	if err != nil {
		log.Println(err)
		return false
	}
	c2, err := redis.Dial("tcp", *connection2)
	if err != nil {
		log.Println("Unable to connect to redis mirror:", err.Error())
	}

	conn.SetContext(RedisSettings{ conns: map[int]redis.Conn{ 1: c1, 2: c2} })
	return true
}

func redisClose(conn redcon.Conn, err error) {
	log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
}

// This function starts as a goroutine. A concurrent process.
// mirrorDo loops forever and run the queued commands against the mirror redis
func mirrorDo() {
	for {
		mirrorDo := <-mirrorDoQueue
		if mirrorDo.rdb == nil {
			log.Println("No mirror connection, skipping", mirrorDo.cmdArgs)
			continue
		}
		_, errMirror := mirrorDo.rdb.Do(mirrorDo.cmdArgs[0].(string), mirrorDo.cmdArgs[1:]...)
		if errMirror != nil {
			log.Println("Mirror Do failed:", errMirror.Error())
		}
	}
}
