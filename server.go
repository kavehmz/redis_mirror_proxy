package main

import (
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/redcon"
)

type RedisSettings struct {
	redisMain   redis.Conn
	redisMirror redis.Conn
}

type rdbDo struct {
	rdb     redis.Conn
	cmdArgs []interface{}
}

// This function will process the incoming messages from client. It will act as a multiplexer.
// To get more information refer to:
// https://redis.io/docs/reference/protocol-spec
// https://pkg.go.dev/github.com/gomodule/redigo/redis#hdr-Executing_Commands
func redisCommand(conn redcon.Conn, cmd redcon.Command) {
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		rdbMain := conn.Context().(RedisSettings).redisMain

		cmdArgs := []interface{}{}
		for _, v := range cmd.Args {
			cmdArgs = append(cmdArgs, string(v))
		}
		res, err := rdbMain.Do(cmdArgs[0].(string), cmdArgs[1:]...)
		if err != nil {
			conn.WriteError(err.Error())
			return
		}

		rdbMirror := conn.Context().(RedisSettings).redisMirror
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
		switch v := res.(type) {
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
		return
	case "subscribe", "psubscribe", "publish":
		conn.WriteError("Unsupported command")
	}
}

func printValue(v []interface{}, conn redcon.Conn) {
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
}

func redisConnect(conn redcon.Conn) bool {
	main, err := redis.Dial("tcp", *mainRedis)
	if err != nil {
		log.Println(err)
		return false
	}
	mirror, err := redis.Dial("tcp", *mirrorRedis)
	if err != nil {
		log.Println("Unable to connect to redis mirror:", err.Error())
	}

	conn.SetContext(RedisSettings{
		redisMain:   main,
		redisMirror: mirror,
	})
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
