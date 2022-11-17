package main

import (
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/redcon"
)

type RedisSettings struct {
	redisMain redis.Conn
}

// This function will process the incoming messages from client. It will act as a multiplexer.
// To get more information refer to:
// https://redis.io/docs/reference/protocol-spec
// https://pkg.go.dev/github.com/gomodule/redigo/redis#hdr-Executing_Commands
func redisCommand(conn redcon.Conn, cmd redcon.Command) {
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		rdb := conn.Context().(RedisSettings).redisMain
		cmdArgs := []interface{}{}
		for _, v := range cmd.Args {
			cmdArgs = append(cmdArgs, string(v))
		}
		res, err := rdb.Do(cmdArgs[0].(string), cmdArgs[1:]...)
		if err != nil {
			conn.WriteError(err.Error())
			return
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

	conn.SetContext(RedisSettings{
		redisMain: main,
	})
	return true
}

func redisClose(conn redcon.Conn, err error) {
	log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
}
