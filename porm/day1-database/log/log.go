package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	errlog  = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infolog = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers = []*log.Logger{errlog, infolog}
	mu      sync.Mutex
)

// log方法
var (
	Error  = errlog.Println
	Errorf = errlog.Printf
	Info   = infolog.Println
	Infof  = infolog.Printf
)

// log层级
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// 设置log层级
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errlog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		infolog.SetOutput(ioutil.Discard)
	}
}
