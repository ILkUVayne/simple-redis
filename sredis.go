package main

import (
	"flag"
	"fmt"
	"simple-redis/src"
)

var banner = `
 _______  ___   __   __  _______  ___      _______         ______    _______  ______   ___   _______ 
|       ||   | |  |_|  ||       ||   |    |       |       |    _ |  |       ||      | |   | |       |
|  _____||   | |       ||    _  ||   |    |    ___| ____  |   | ||  |    ___||  _    ||   | |  _____|
| |_____ |   | |       ||   |_| ||   |    |   |___ |____| |   |_||_ |   |___ | | |   ||   | | |_____ 
|_____  ||   | |       ||    ___||   |___ |    ___|       |    __  ||    ___|| |_|   ||   | |_____  |
 _____| ||   | | ||_|| ||   |    |       ||   |___        |   |  | ||   |___ |       ||   |  _____| |
|_______||___| |_|   |_||___|    |_______||_______|       |___|  |_||_______||______| |___| |_______|
`

const (
	VERSION = "0.0.0"
	CONFIG  = "./sredis.conf"
)

func main() {
	// initialization
	fmt.Printf("%s\n\n", banner)
	fmt.Printf("version: %s\n", VERSION)
	confPath := flag.String("c", CONFIG, "config path")
	flag.Parse()
	// load config
	src.SetupConf(*confPath)
	// server start
	src.ServerStart()
}
