// Package src
//
// Lib args provides server-side and client-side parameter parsing
package src

import (
	"flag"
	"unicode"
)

var transChar = map[string]string{
	"n": "\n",
	"r": "\r",
	"t": "\t",
	"b": "\b",
	"a": "\a",
}

// ------------------------------ args tools --------------------------

// return next not space index by index
//
// e.g. s = "hello  world" i = 5, will return 7
func nextLineIdx(line string, i int) int {
	for i < len(line) && unicode.IsSpace(rune(line[i])) {
		i++
	}
	return i
}

func checkTerminated(line, key string, i int) (status int) {
	// closing quote must be followed by a space or nothing at all
	if string(line[i]) == key {
		if i+1 < len(line) && !unicode.IsSpace(rune(line[i+1])) {
			return SPA_TERMINATED
		}
		return SPA_DONE
	}
	// unterminated quotes
	if i+1 == len(line) {
		return SPA_TERMINATED
	}
	return SPA_CONTINUE
}

func normalHandle(line string, i int) (string, int, int) {
	current := ""
	for ; i < len(line); i++ {
		if unicode.IsSpace(rune(line[i])) {
			break
		}
		current += string(line[i])
	}
	return current, i, SPA_DONE
}

func quotesHandle(line string, i int) (string, int, int) {
	current := ""
	for ; i < len(line); i++ {
		// e.g. \r \n \" and so on
		if string(line[i]) == "\\" && i+1 < len(line) {
			i++
			tc, ok := transChar[string(line[i])]
			if !ok {
				current += string(line[i])
				continue
			}
			current += tc
			continue
		}
		// closing quote must be followed by a space or nothing at all
		if status := checkTerminated(line, "\"", i); status != SPA_CONTINUE {
			return current, i, status
		}
		current += string(line[i])
		continue
	}
	return current, i, SPA_DONE
}

func singleQuotesHandle(line string, i int) (string, int, int) {
	current := ""
	for ; i < len(line); i++ {
		// e.g. \r \n \" and so on
		if string(line[i]) == "\\" && i+1 < len(line) && string(line[i+1]) == "'" {
			current += "'"
			i++
			continue
		}
		// closing quote must be followed by a space or nothing at all
		if status := checkTerminated(line, "'", i); status != SPA_CONTINUE {
			return current, i, status
		}
		current += string(line[i])
		continue
	}
	return current, i, SPA_DONE
}

func splitArgs(line string) []string {
	if len(line) == 0 {
		return nil
	}
	var args []string
	// skip space
	i := nextLineIdx(line, 0)
	for i < len(line) {
		current, status := "", 0
		switch string(line[i]) {
		case "\"":
			current, i, status = quotesHandle(line, i+1)
		case "'":
			current, i, status = singleQuotesHandle(line, i+1)
		default:
			current, i, status = normalHandle(line, i)
		}
		if status == SPA_TERMINATED {
			return nil
		}
		args = append(args, current)
		i = nextLineIdx(line, i+1)
	}
	return args
}

// ----------------------------- server args -------------------------

// server args
type serverArgs struct {
	confPath string // config file path
}

var ServerArgs serverArgs

func ParseServerArgs() {
	flag.StringVar(&ServerArgs.confPath, "c", CONFIG, "config path")
	flag.Parse()
}

// ------------------------------- cli args ---------------------------

type cliArgs struct {
	hostIp string
	port   int
	prompt string // client cli prompt
}

var CliArgs cliArgs

func ParseCliArgs() {
	flag.StringVar(&CliArgs.hostIp, "host", "127.0.0.1", "Server hostname")
	flag.IntVar(&CliArgs.port, "p", DEFAULT_PORT, "Server port")
	flag.Parse()
}
