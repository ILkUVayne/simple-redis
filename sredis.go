package main

import (
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
)

func main() {
	//utils.SetLevel(utils.ErrorLevel)
	// initialization
	fmt.Printf("%s\n\n", banner)
	fmt.Printf("version: %s\n", VERSION)
	// parse args
	src.ParseServerArgs()
	// server start
	src.ServerStart()
}
