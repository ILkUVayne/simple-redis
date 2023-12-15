package main

import (
	"flag"
	"simple-redis/src"
)

func main() {
	// parse args
	src.ParseCliArgs()
	// start cli
	src.CliStart(flag.Args())
}
