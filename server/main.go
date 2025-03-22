package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nthnn/ura/db"
	"github.com/nthnn/ura/logger"
	"github.com/nthnn/ura/mux"
)

type Config struct {
	Address  string `json:"address"`
	Port     int16  `json:"port"`
	Database string `json:"database"`
	Root     struct {
		Base string `json:"base"`
		Dir  string `json:"dir"`
	} `json:"root"`
}

var (
	database *sql.DB
	config   Config
)

func loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		return err
	}
	return nil
}

func initServer() {
	var err error
	database, err = db.Initialize(config.Database)
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	mux.Initialize(config.Address, config.Port)
	logger.Info("Starting server on %s:%d.", config.Address, config.Port)

	mux.InitializeEntryPoints(database)
	logger.Info("Initialized server entry points!")

	mux.RootDirectory(config.Root.Base, config.Root.Dir)
	logger.Info("Serving static files from %s/%s.", config.Root.Base, config.Root.Dir)
}

func runServer() {
	if err := mux.Start(); err != nil {
		logger.Error("Server failed to start on %s:%d: %v", config.Address, config.Port, err)
		os.Exit(1)
	}
}

func panicRecovery(done chan bool) {
	if <-done {
		if database != nil {
			database.Close()
		}
		os.Exit(0)
	}
	if r := recover(); r != nil {
		defer panicRecovery(done)
		logger.Error("Panic occurred: %v", r)
		logger.Log("Recovering from panic and restarting server...")
		runServer()
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	defer panicRecovery(done)

	exe, err := os.Executable()
	if err != nil {
		logger.Error("Error getting executable path: %s", err.Error())
		os.Exit(1)
	}

	configPath := filepath.Join(filepath.Dir(exe), "config.json")
	if err := loadConfig(configPath); err != nil {
		logger.Error("Error loading config.json: %s", err.Error())
		os.Exit(1)
	}

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
