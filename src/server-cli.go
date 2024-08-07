package src

import (
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/ILkUVayne/utlis-go/v2/cli"
	"github.com/ILkUVayne/utlis-go/v2/str"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"simple-redis/utils"
	"strings"
)

var context *sRedisContext

/*------------------------------------------------------------------------------
 * Networking / parsing
 *--------------------------------------------------------------------------- */

func sRedisConnect() *sRedisContext {
	c := new(sRedisContext)
	host, err := str.IPStrToHost(CliArgs.hostIp)
	if err != nil {
		ulog.Error(err)
	}
	c.fd = Connect(host, CliArgs.port)
	return c
}

func cliConnect(force int) int {
	if context == nil || force > 0 {
		context = sRedisConnect()

		// Do AUTH
	}
	return CLI_OK
}

func cliSendCommand(args []string) int {
	if context == nil {
		return CLI_ERR
	}
	var reply sRedisReply
	sRedisAppendCommandArg(context, args)
	sRedisGetReply(context, &reply)
	context.reader = &reply
	printPrompt()
	return CLI_OK
}

/*------------------------------------------------------------------------------
 * Utility functions
 *--------------------------------------------------------------------------- */

func cliRefreshPrompt() {
	CliArgs.prompt = fmt.Sprintf("%s:%d> ", CliArgs.hostIp, CliArgs.port)
}

func cliDisplayPrompt() string {
	if context == nil {
		return "not connected> "
	}
	return CliArgs.prompt
}

func cliInputLine() string {
	s, err := linenoise.Line(cliDisplayPrompt())
	if err != nil {
		// KillSignalError
		utils.Exit(0)
	}
	return s
}

func printPrompt() {
	if context.err != nil {
		fmt.Printf("(error) ERR: " + context.err.Error())
	}
	reader := context.reader.(*sRedisReply)
	if reader.fStr != "" {
		fmt.Println(reader.fStr)
		return
	}
	fmt.Println(reader.str)
}

/*------------------------------------------------------------------------------
 * User interface
 *--------------------------------------------------------------------------- */

func parseOptions() {

}

func repl() {
	history, hf := false, ""
	if cli.Isatty() {
		history, hf = true, utils.HistoryFile(REDIS_CLI_HISTFILE_DEFAULT)
		_ = linenoise.LoadHistory(hf)
	}

	cliRefreshPrompt()
	for {
		s := cliInputLine()
		if len(s) == 0 {
			fmt.Println("Invalid argument(s)")
			continue
		}
		if history {
			_, _ = linenoise.AddHistory(s), linenoise.SaveHistory(hf)
		}
		fields := strings.Fields(s)
		if fields[0] == "quit" || fields[0] == "exit" {
			utils.Exit(0)
		}
		// cliSendCommand
		if cliSendCommand(fields) != CLI_OK {
			cliConnect(1)
			if cliSendCommand(fields) != CLI_OK {
				ulog.Error("simple-redis cli: cliSendCommand error")
			}
		}
	}
}

func noninteractive(args []string) {
	ulog.Info(args)
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
		utils.Exit(1)
	}
	noninteractive(args)
}
