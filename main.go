package main

import (
	"Users/database"
	"Users/handlers"
	"Users/middleware"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ------------------ MAIN ------------------

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := database.ConnectDB()
	//
	h := &handlers.Handler{Client: client}
	//logger using slog to log in json format

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {

			logger.Error("Error disconnecting from mongoDb", "error", err)
		}
	}()

	//using a server mux to map the requests to the handlers
	mux := http.NewServeMux()

	//applying method check middleware to the mux handlers
	mux.Handle("/api/create-user", middleware.MethodChecker([]string{http.MethodPost}, http.HandlerFunc(h.CreateUser)))
	mux.Handle("/api/update-user/{id}", middleware.MethodChecker([]string{http.MethodPut}, http.HandlerFunc(h.UpdateUser)))
	mux.Handle("/api/users", middleware.MethodChecker([]string{http.MethodGet}, http.HandlerFunc(h.FetchAllUsers)))
	mux.Handle("/api/delete-user/{id}", middleware.MethodChecker([]string{http.MethodDelete}, http.HandlerFunc(h.DeleteUser)))
	mux.Handle("/api/update-status/{id}", middleware.MethodChecker([]string{http.MethodPut}, http.HandlerFunc(h.UpdateStatus)))
	mux.Handle("/api/emails", middleware.MethodChecker([]string{http.MethodGet}, http.HandlerFunc(h.FetchAllEmails)))

	//Wrapping the mux around the panic middleware

	handlerforPanicRecovery := middleware.PanicMiddleware(logger)(mux)
	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerforPanicRecovery,
	}
	logger.Info("Server started on port 8080")
	go func() {

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server has been shutdown...")
	}
	logger.Info("Server exited...")
}
