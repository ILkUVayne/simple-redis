// Package src
//
// Lib args provides server-side and client-side parameter parsing
package src

import (
	"flag"
	"unicode"
)

// 需要转义的字符串map
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

// check double quotes string or single quotes string is Terminated,closing quote
// must be followed by a space or nothing at all.
//
// return SPA_CONTINUE, if string is not complete.
// return SPA_DONE, if string is complete and valid.
// return SPA_TERMINATED, if string is invalid.
func checkTerminated(line, key string, i int) (status int) {
	if string(line[i]) == key {
		if i+1 < len(line) && !unicode.IsSpace(rune(line[i+1])) {
			return SPA_TERMINATED
		}
		return SPA_DONE
	}
	// unterminated quotes. e.g. "hello   or 'hello
	if i+1 == len(line) {
		return SPA_TERMINATED
	}
	return SPA_CONTINUE
}

// 没有单引号或双引号包裹的普通字符串参数
func normalArgs(line string, i int) (string, int, int) {
	current := ""
	for ; i < len(line); i++ {
		if unicode.IsSpace(rune(line[i])) {
			break
		}
		current += string(line[i])
	}
	return current, i, SPA_DONE
}

// 双引号包裹的字符串参数
func quotesArgs(line string, i int) (string, int, int) {
	current := ""
	for i++; i < len(line); i++ {
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

// 单引号包裹的字符串参数
func singleQuotesArgs(line string, i int) (string, int, int) {
	current := ""
	for i++; i < len(line); i++ {
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

// 解析拆分字符串命令
func splitArgs(line string) (args []string) {
	if len(line) == 0 {
		return nil
	}
	// skip space
	i := nextLineIdx(line, 0)
	for i < len(line) {
		current, status := "", 0
		current, i, status = splitArgsHandle(string(line[i]), line, i)
		if status == SPA_TERMINATED {
			return nil
		}
		args = append(args, current)
		i = nextLineIdx(line, i+1)
	}
	return
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
	auth   string // auth password, default ""
}

var CliArgs cliArgs

func ParseCliArgs() {
	flag.StringVar(&CliArgs.hostIp, "host", "127.0.0.1", "Server hostname")
	flag.IntVar(&CliArgs.port, "p", 6379, "Server port")
	flag.StringVar(&CliArgs.auth, "a", "", "auth password")
	flag.Parse()
}
