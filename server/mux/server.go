package mux

import (
	"context"
	"database/sql"
	"net/http"
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

		logger.Log(
			"Request: %s %s | Remote Address: %s | Duration: %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			duration,
		)
	})
}

func Initialize(port int16) {
	muxServer = http.NewServeMux()
	httpServer = &http.Server{
		Addr:    "localhost:" + strconv.Itoa(int(port)),
		Handler: loggingMiddleware(muxServer),
	}
}

func InitializeEntryPoints(db *sql.DB) {
	addEntryPoint("/api/user/create", db, handler.UserCreate)
	addEntryPoint("/api/user/delete", db, handler.UserDelete)
	addEntryPoint("/api/user/login", db, handler.UserLogin)
	addEntryPoint("/api/user/logout", db, handler.UserLogout)
	addEntryPoint("/api/user/notifications", db, handler.UserFetchNotifications)

	addEntryPoint("/api/loan/request", db, handler.LoanRequest)
	addEntryPoint("/api/loan/accept", db, handler.LoanAccept)
	addEntryPoint("/api/loan/reject", db, handler.LoanReject)

	addEntryPoint("/api/payment/transaction", db, handler.PaymentTransaction)
	addEntryPoint("/api/payment/request", db, handler.PaymentRequest)

	addEntryPoint("/api/refund/request", db, handler.RefundRequest)
	addEntryPoint("/api/refund/reject", db, handler.RefundReject)
	addEntryPoint("/api/refund/process", db, handler.RefundProcess)

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

func RootDirectory(folderName string) {
	muxServer.Handle(
		"/",
		http.FileServer(http.Dir("./"+folderName)),
	)
}

func Start() error {
	return httpServer.ListenAndServe()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down HTTP server: %s", err.Error())
	}
}
