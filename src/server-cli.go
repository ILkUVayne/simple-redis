package src

import (
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	"simple-redis/utils"
	"strings"
)

const (
	CLI_OK  = 0
	CLI_ERR = 1

	REDIS_CLI_HISTFILE_DEFAULT = ".srediscli_history"
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

func getHome() string {
	str, err := utils.Home()
	if err != nil {
		utils.Error(err)
	}
	return str
}

func historyFile(file string) string {
	return getHome() + "/" + file
}

/*------------------------------------------------------------------------------
 * User interface
 *--------------------------------------------------------------------------- */

func parseOptions() {

}

func repl() {
	history := false
	historyfile := ""

	if utils.Isatty() {
		history = true
		historyfile = historyFile(REDIS_CLI_HISTFILE_DEFAULT)
		_ = linenoise.LoadHistory(historyfile)
	}

	cliRefreshPrompt()
	for {
		var str string
		var err error
		if context == nil {
			str, err = linenoise.Line("not connected> ")
		} else {
			str, err = linenoise.Line(CliArgs.prompt)
		}
		if err != nil {
			break
		}

		if history {
			_ = linenoise.AddHistory(str)
			_ = linenoise.SaveHistory(historyfile)
		}
		if len(str) == 0 {
			fmt.Println("Invalid argument(s)")
			continue
		}
		fields := strings.Fields(str)
		if fields[0] == "quit" || fields[0] == "exit" {
			utils.Exit(0)
		}
		// cliSendCommand
	}
	utils.Exit(0)
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
		utils.Exit(1)
	}
	noninteractive(args)
}
