package main

import "simple-redis/src"

func main() {
	// parse args
	src.ParseCliArgs()
	// start cli
	src.CliStart()
}
