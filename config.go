package main

import (
	"fmt"

	"gopkg.in/ini.v1"
)

func config() (*string, *string, *string) {
	fmt.Println("Load config")
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
	}

	var main string = cfg.Section("main").Key("server").String() + ":" + cfg.Section("main").Key("port").String()
	var mirror string = cfg.Section("mirror").Key("server").String() + ":" + cfg.Section("mirror").Key("port").String()
	var proxy string = cfg.Section("proxy").Key("server").String() + ":" + cfg.Section("proxy").Key("port").String()

	return &main, &mirror, &proxy
}

