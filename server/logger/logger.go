package logger

import (
	"fmt"
	"sync"
	"time"
)

var logMutex sync.Mutex

func Log(data string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fmt.Printf(
		"%s\t\u001b[44m  LOG  \u001b[0m\t :: ",
		time.Now().Format(time.RFC3339Nano),
	)

	fmt.Printf(data, args...)
	fmt.Println()
}

func Info(data string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fmt.Printf(
		"%s\t\u001b[42m INFO  \u001b[0m\t :: ",
		time.Now().Format(time.RFC3339Nano),
	)

	fmt.Printf(data, args...)
	fmt.Println()
}

func Error(data string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fmt.Printf(
		"%s\t\u001b[41m ERROR \u001b[0m\t :: ",
		time.Now().Format(time.RFC3339Nano),
	)

	fmt.Printf(data, args...)
	fmt.Println()
}
