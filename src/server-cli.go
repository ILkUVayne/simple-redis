package src

import (
	"errors"
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/ILkUVayne/utlis-go/v2/cli"
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
	c.fd, c.err = Connect(ipStrToHost(CliArgs.hostIp), CliArgs.port)
	return c
}

// check connect available, return false if unconnected.
func checkConnected() bool {
	return context != nil
}

// return false if you have error.
func checkError() bool {
	return context.err == nil
}

// client auth
func cliAuth() {
	if CliArgs.auth != "" {
		cliSendCommand1([]string{"auth", CliArgs.auth})
	}
}

// try to connect server
func cliConnect(force int) bool {
	if context == nil || force > 0 {
		context = sRedisConnect()
	}

	// 连接出错，返回
	if !checkError() {
		return false
	}

	// Do AUTH
	cliAuth()

	return checkError()
}

// send cli command to server and get reply
func cliSendCommand1(args []string) bool {
	var reply sRedisReply
	sRedisAppendCommandArg(context, args)
	sRedisGetReply(context, &reply)
	context.reader = &reply
	return checkError()
}

// send cli command to server and get reply,will retry 1 time if context is nil
func cliSendCommand(args []string) {
	// 连接不存在或者重连失败，直接返回
	if !checkConnected() && !cliConnect(1) {
		return
	}
	// 连接存在，发送请求
	cliSendCommand1(args)
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
		fmt.Printf("Error: Server closed the connection or restart, try again please\r\n")
	default:
		fmt.Printf("(error) ERR: " + context.err.Error() + "\r\n")
	}
	context = nil
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
		printPrompt()
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
	if !cliConnect(0) {
		os.Exit(1)
	}
	noninteractive(args)
}
