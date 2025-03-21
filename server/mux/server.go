package mux

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nthnn/ura/handler"
	"github.com/nthnn/ura/logger"
	"github.com/nthnn/ura/util"
)

var (
	muxServer  *http.ServeMux
	httpServer *http.Server
)

func addEntryPoint(
	path string,
	db *sql.DB,
	callback func(*sql.DB) func(http.ResponseWriter, *http.Request),
) {
	muxServer.Handle(
		path,
		util.RateLimit(
			http.HandlerFunc(callback(db)),
		),
	)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		logger.Log(
			"Request: %s %s | Remote Address: %s | Duration: %v",
			r.Method,
			r.URL.Path,
			clientIP,
			duration,
		)
	})
}

func Initialize(bindAddr string, port int16) {
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}

	muxServer = http.NewServeMux()
	addr := bindAddr + ":" + strconv.Itoa(int(port))

	httpServer = &http.Server{
		Addr:    addr,
		Handler: loggingMiddleware(muxServer),
	}
}

func InitializeEntryPoints(db *sql.DB) {
	addEntryPoint("/api/user/create", db, handler.UserCreate)
	addEntryPoint("/api/user/delete", db, handler.UserDelete)
	addEntryPoint("/api/user/login", db, handler.UserLogin)
	addEntryPoint("/api/user/logout", db, handler.UserLogout)

	addEntryPoint("/api/payment/send", db, handler.PaymentProcess)
	addEntryPoint("/api/payment/request", db, handler.PaymentRequest)

	addEntryPoint("/api/withdraw", db, handler.Withdraw)
	addEntryPoint("/api/cashin", db, handler.CashIn)

	muxServer.Handle(
		"/api/user/session",
		http.HandlerFunc(handler.ValidateSession(db)),
	)

	muxServer.Handle(
		"/api/user/info",
		http.HandlerFunc(handler.UserFetchInfo(db)),
	)
}

func RootDirectory(baseDir, folderName string) {
	safeFolder := filepath.Clean(folderName)
	absPath := filepath.Join(baseDir, safeFolder)

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		logger.Error("Error obtaining absolute path for baseDir: %s", err.Error())
		return
	}

	absPath, err = filepath.Abs(absPath)
	if err != nil {
		logger.Error("Error obtaining absolute path for folder: %s", err.Error())
		return
	}

	if !filepath.HasPrefix(absPath, baseAbs) {
		logger.Error("Attempted directory traversal in static file serving: %s", folderName)
		return
	}

	muxServer.Handle(
		"/",
		http.FileServer(http.Dir(absPath)),
	)
}

func Start() error {
	return httpServer.ListenAndServe()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down HTTP server: %s", err.Error())
	}
}
