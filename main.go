package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insomnius/wallet-event-loop/agregation"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/handler"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	fmt.Println("Starting e-wallet services...")

	fmt.Println("Starting e-wallet databases...")
	dbInstance := db.NewInstance()

	// Starting database instance
	go func() {
		dbInstance.Start()
	}()

	dbInstance.CreateTable("users")
	dbInstance.CreateTable("user_tokens")
	dbInstance.CreateTable("wallets")
	dbInstance.CreateTable("transactions")

	e := echo.New()
	e.Use(middleware.Logger())

	e.Use(middleware.Recover())

	walletRepo := repository.NewWallet(dbInstance)
	userRepo := repository.NewUser(dbInstance)
	userTokenRepo := repository.NewUserToken(dbInstance)
	// mutationRepo := repository.NewMutation(dbInstance)

	authAggregator := agregation.NewAuthorization(
		walletRepo,
		userRepo,
		userTokenRepo,
		dbInstance,
	)

	e.POST("/users", handler.UserRegister(authAggregator))
	e.POST("/users/signin", handler.UserSignin(authAggregator))

	go func() {
		port := "8000"
		if os.Getenv("PORT") != "" {
			port = os.Getenv("PORT")
		}

		fmt.Println("Starting http server on port:", port)

		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error shutting down the server. Error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Printf("\nShutting down the server...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		fmt.Println("Error shutting down the server. Error:", err)
	}

	dbInstance.Close()
	fmt.Println("Closing e-wallet database...")
}
