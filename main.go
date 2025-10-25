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
	//logger using slog to log in json format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := database.ConnectDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {

			logger.Error("Error disconnecting from mongoDb", "error", err)
		}
	}()

	//using a server mux to map the requests to the handlers
	mux := http.NewServeMux()

	//applying method check middleware to the mux handlers

	createUser := middleware.MethodChecker([]string{http.MethodPost}, handlers.CreateUser)
	mux.Handle("/api/create-user", createUser)
	updateUser := middleware.MethodChecker([]string{http.MethodPut}, handlers.UpdateUser)
	mux.Handle("/api/update-user/{id}", updateUser)
	fetchAllUsers := methodChecker([]string{http.MethodGet}, http.HandlerFunc(fetchAllUsers))
	mux.Handle("/api/users", fetchAllUsers)
	deleteUser := methodChecker([]string{http.MethodDelete}, http.HandlerFunc(deleteUser))
	mux.Handle("/api/delete-user/{id}", deleteUser)
	updateStatus := methodChecker([]string{http.MethodPut}, http.HandlerFunc(updateStatus))
	mux.Handle("/api/update-Status/", updateStatus)
	fetchAllEmails := methodChecker([]string{http.MethodGet}, http.HandlerFunc(fetchAllEmails))
	mux.Handle("/api/emails", fetchAllEmails)

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
