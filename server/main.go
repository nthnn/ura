package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nthnn/ura/db"
	"github.com/nthnn/ura/logger"
	"github.com/nthnn/ura/mux"
)

var (
	port    = int16(5173)
	retries = 0
)

func initServer() {
	db, err := db.Initialize()
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	mux.Initialize(port)
	logger.Info("Starting server on port %d.", port)

	mux.InitializeEntryPoints(db)
	logger.Info("Initialized server entry points!")

	mux.RootDirectory("public")
	logger.Info("Added /public/ directory to accessible client-side paths.")
}

func runServer() {
	if err := mux.Start(); err != nil {
		if retries < 30 {
			retries++
			panic("Cannot start server on port " + strconv.Itoa(int(port)) + "!")
		} else {
			logger.Error("Server failed to start after 30 attempts.")
			os.Exit(0)
		}
	}
}

func panicRecovery(done chan bool) {
	if <-done {
		os.Exit(0)
	}

	if r := recover(); r != nil {
		defer panicRecovery(done)
		logger.Error("%v", r)

		logger.Log("Recovering from panic...")
		runServer()
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	defer panicRecovery(done)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Info("Received signal: %s", sig.String())

		mux.Stop()
		done <- true
	}()

	initServer()
	runServer()
}
