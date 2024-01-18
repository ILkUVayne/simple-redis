package utils

import (
	"io"
	"log"
	"os"
	"sync"
)

const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags)
	loggers  = []*log.Logger{errorLog, infoLog}
	mux      sync.Mutex
)

// Error ErrorF 会阻止defer执行
// ErrorP ErrorPf 不会会阻止defer执行，但需要手动return
var (
	Error   = errorLog.Fatal
	ErrorF  = errorLog.Fatalf
	ErrorP  = errorLog.Println
	ErrorPf = errorLog.Printf
	Info    = infoLog.Println
	InfoF   = infoLog.Printf
)

func SetLevel(level int) {
	mux.Lock()
	defer mux.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(io.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(io.Discard)
	}
}
