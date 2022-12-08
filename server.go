package main

import (
	"log"
	"strings"
	"sync/atomic"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/redcon"
)

type RedisSettings struct {
	conns     map[int]redis.Conn
	replyChan chan reply
}

type reply struct {
	res interface{}
	err error
}

var ps redcon.PubSub
var mode atomic.Int32

// This function will process the incoming messages from client. It will act as a multiplexer.
// To get more information refer to:
// https://redis.io/docs/reference/protocol-spec
// https://pkg.go.dev/github.com/gomodule/redigo/redis#hdr-Executing_Commands
func redisCommand(conn redcon.Conn, cmd redcon.Command) {
	cmdArgs := []interface{}{}
	for _, v := range cmd.Args {
		cmdArgs = append(cmdArgs, string(v))
	}

	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		mainQueue <- mainCmd{conn: conn, cmdArgs: cmdArgs}
		rep := <-conn.Context().(RedisSettings).replyChan
		res := rep.res
		err := rep.err
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		if res == nil {
			conn.WriteNull()
			return
		}
		respond(rep.res, conn)
	// return the number of queue commands in mirror queue
	case "qlen":
		conn.WriteInt(len(mirrorQueue))
	case "switch":
		switchConnections()
		conn.WriteString("Switch is done")
	case "subscribe", "psubscribe":
		// subscribe to both redises to eliminte the need of any action in case of switch
		err := pickMain(conn).Send(cmdArgs[0].(string), cmdArgs[1:]...)
		if err != nil {
			log.Println("Unable to send the subscribe command", err)
			conn.WriteError(err.Error())
			return
		}
		err = pickMain(conn).Flush()
		if err != nil {
			log.Println("Unable to flush after subscribe", err)
			conn.WriteError(err.Error())
			return
		}

		err = pickMirror(conn).Send(cmdArgs[0].(string), cmdArgs[1:]...)
		if err != nil {
			log.Println("Unable to send the subscribe command in mirror", err)
		}
		err = pickMirror(conn).Flush()
		if err != nil {
			log.Println("Unable to flush after subscribe in mirorr", err)

		}

		subCmd := strings.ToLower(string(cmd.Args[0]))
		for i := 1; i < len(cmd.Args); i++ {
			if subCmd == "psubscribe" {
				ps.Psubscribe(conn, string(cmd.Args[i]))
			} else {
				ps.Subscribe(conn, string(cmd.Args[i]))
			}
		}

		// receive main
		go func(conn redcon.Conn) {
			for {
				res, err := pickMain(conn).Receive()
				if err != nil {
					log.Println("Error in receiving the next message for subscription", err)
					conn.WriteError(err.Error())
					return
				}
				msgType := string(res.([]interface{})[0].([]byte))
				if msgType == "message" {
					// In the case of "message", the format is:
					// [message] [channel] [value]
					// Like: message mychannelname publishedvalue
					ps.Publish(string(res.([]interface{})[1].([]byte)), string(res.([]interface{})[2].([]byte)))
				}
				if msgType == "pmessage" {
					// In the case of "pmessage", the format is:
					// [message] [channel_pattern] [channel] [value]
					// Like: message mychannelname? mychannelname1 publishedvalue
					ps.Publish(string(res.([]interface{})[2].([]byte)), string(res.([]interface{})[3].([]byte)))
				}
			}
		}(conn)
		// receive mirror
		go func(conn redcon.Conn) {
			for {
				res, err := pickMirror(conn).Receive()
				if err != nil {
					log.Println("Error in receiving the next message for subscription", err)
					conn.WriteError(err.Error())
					return
				}
				msgType := string(res.([]interface{})[0].([]byte))
				if msgType == "message" {
					// In the case of "message", the format is:
					// [message] [channel] [value]
					// Like: message mychannelname publishedvalue
					ps.Publish(string(res.([]interface{})[1].([]byte)), string(res.([]interface{})[2].([]byte)))
				}
				if msgType == "pmessage" {
					// In the case of "pmessage", the format is:
					// [message] [channel_pattern] [channel] [value]
					// Like: message mychannelname? mychannelname1 publishedvalue
					ps.Publish(string(res.([]interface{})[2].([]byte)), string(res.([]interface{})[3].([]byte)))
				}
			}
		}(conn)
	}
}

func pickMain(conn redcon.Conn) redis.Conn {
	if mode.Load()%2 == 0 {
		return conn.Context().(RedisSettings).conns[1]
	}
	return conn.Context().(RedisSettings).conns[2]
}

func pickMirror(conn redcon.Conn) redis.Conn {
	if mode.Load()%2 == 0 {
		return conn.Context().(RedisSettings).conns[2]
	}
	return conn.Context().(RedisSettings).conns[1]
}

func respond(res interface{}, conn redcon.Conn) {
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
				respond(v, conn)
			default:
				log.Printf("This is an unknow type! %T", v)
			}
		}
	default:
		log.Printf("This is an unknow type! %T", v)
	}
}

func redisConnect(conn redcon.Conn) bool {
	c1, err := redis.Dial("tcp", *redisAddr1)
	if err != nil {
		log.Println(err)
		return false
	}
	c2, err := redis.Dial("tcp", *redisAddr2)
	if err != nil {
		log.Println("Unable to connect to redis mirror:", err.Error())
	}

	conn.SetContext(RedisSettings{conns: map[int]redis.Conn{1: c1, 2: c2}, replyChan: make(chan reply)})
	return true
}

func redisClose(conn redcon.Conn, err error) {
	log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
}
