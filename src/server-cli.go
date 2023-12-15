package src

import (
	"fmt"
	"os"
	"simple-redis/utils"
	"time"
)

const (
	CLI_OK  = 0
	CLI_ERR = 1
)

type sRedisContext struct {
	fd     int
	obuf   []byte
	reader []byte
}

var context *sRedisContext

func sRedisContextInit() *sRedisContext {
	c := new(sRedisContext)
	return c
}

/*------------------------------------------------------------------------------
 * Networking / parsing
 *--------------------------------------------------------------------------- */

func sRedisConnect() *sRedisContext {
	c := sRedisContextInit()
	c.fd = Connect(utils.StrToHost(CliArgs.hostIp), CliArgs.port)
	return c
}

func cliConnect(force int) int {
	if context == nil || force > 0 {
		context = sRedisConnect()

		// Do AUTH
	}
	return CLI_OK
}

/*------------------------------------------------------------------------------
 * Utility functions
 *--------------------------------------------------------------------------- */

func cliRefreshPrompt() {
	CliArgs.prompt = fmt.Sprintf("%s:%d> ", CliArgs.hostIp, CliArgs.port)
}

/*------------------------------------------------------------------------------
 * User interface
 *--------------------------------------------------------------------------- */

func parseOptions() {

}

func repl() {
	cliRefreshPrompt()
	fmt.Print(CliArgs.prompt)
	time.Sleep(time.Second * 3)
	//for {
	//	var args []byte
	//	if string(args[0]) == "quit" || string(args[0]) == "exit" {
	//		os.Exit(0)
	//	}
	//}
}

func noninteractive(args []string) {
	//
}

func CliStart(args []string) {
	parseOptions()

	// Start interactive mode when no command is provided
	if len(args) == 0 {
		cliConnect(0)
		repl()
	}

	// Otherwise, we have some arguments to execute
	if cliConnect(0) == CLI_ERR {
		os.Exit(1)
	}
	noninteractive(args)
}
