package src

import (
	"errors"
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/ILkUVayne/utlis-go/v2/cli"
	"github.com/ILkUVayne/utlis-go/v2/str"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"golang.org/x/sys/unix"
	"os"
)

var context *sRedisContext

/*------------------------------------------------------------------------------
 * Networking / parsing
 *--------------------------------------------------------------------------- */

// connect server
func sRedisConnect() *sRedisContext {
	c := new(sRedisContext)
	host, err := str.IPStrToHost(CliArgs.hostIp)
	if err != nil {
		ulog.Error(err)
	}
	c.fd, c.err = Connect(host, CliArgs.port)
	return c
}

// try to connect server
func cliConnect(force int) int {
	if context == nil || force > 0 {
		context = sRedisConnect()

		// Do AUTH
	}
	return CLI_OK
}

// send cli command to server and get reply
func cliSendCommand1(args []string) int {
	if context == nil || context.err != nil {
		return CLI_ERR
	}
	var reply sRedisReply
	sRedisAppendCommandArg(context, args)
	sRedisGetReply(context, &reply)
	context.reader = &reply
	return CLI_OK
}

// send cli command to server and get reply,will retry 1 time if context is nil
func cliSendCommand(args []string) {
	if cliSendCommand1(args) != CLI_OK {
		cliConnect(1)
		cliSendCommand1(args)
	}
	printPrompt()
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
		os.Exit(0)
	}
	return s
}

// return false if err is nil
func printErrorPrompt() bool {
	if context.err == nil {
		return false
	}
	switch {
	case errors.Is(context.err, unix.ECONNREFUSED):
		fmt.Printf("Could not connect to simple-redis at %s:%d: Connection refused\r\n", CliArgs.hostIp, CliArgs.port)
	case errors.Is(context.err, CONN_DISCONNECTED):
		fmt.Printf("Error: Server closed the connection\r\n")
		context = nil
	default:
		fmt.Printf("(error) ERR: " + context.err.Error() + "\r\n")
	}
	return true
}

// print server reply
func printPrompt() {
	if printErrorPrompt() {
		return
	}
	reader := context.reader.(*sRedisReply)
	if reader.fStr != "" {
		fmt.Println(reader.fStr)
		return
	}
	fmt.Println(reader.str)
}

// Trim args
//
// e.g. "set" '"name"' "'aaa'" => set "name" 'aaa'
func cliTrimArgs(line string) []string {
	return splitArgs(line)
}

/*------------------------------------------------------------------------------
 * User interface
 *--------------------------------------------------------------------------- */

func parseOptions() {

}

// interactive mode loop
func repl() {
	history, hf := false, ""
	if cli.Isatty() {
		history, hf = true, HistoryFile(REDIS_CLI_HISTFILE_DEFAULT)
		_ = linenoise.LoadHistory(hf)
	}

	cliRefreshPrompt()
	for {
		s := cliInputLine()
		fields := cliTrimArgs(s)
		if len(s) == 0 || fields == nil {
			fmt.Println("Invalid argument(s)")
			continue
		}
		if history {
			_, _ = linenoise.AddHistory(s), linenoise.SaveHistory(hf)
		}
		if fields[0] == "quit" || fields[0] == "exit" {
			os.Exit(0)
		}
		// cliSendCommand
		cliSendCommand(fields)
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
		os.Exit(1)
	}
	noninteractive(args)
}
