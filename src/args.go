package src

import "flag"

// ----------------------------- server args -------------------------
type serverArgs struct {
	confPath string
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
	prompt string
}

var CliArgs cliArgs

func ParseCliArgs() {
	flag.StringVar(&CliArgs.hostIp, "host", "127.0.0.1", "Server hostname")
	flag.IntVar(&CliArgs.port, "p", 6379, "Server port")
	flag.Parse()
}
