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
	"github.com/insomnius/wallet-event-loop/handler/middleware"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
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
	dbInstance.CreateTable("mutations")

	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	walletRepo := repository.NewWallet(dbInstance)
	userRepo := repository.NewUser(dbInstance)
	userTokenRepo := repository.NewUserToken(dbInstance)
	mutationRepo := repository.NewMutation(dbInstance)

	authAggregator := agregation.NewAuthorization(
		walletRepo,
		userRepo,
		userTokenRepo,
		dbInstance,
	)

	trxAggregator := agregation.NewTransaction(
		walletRepo,
		userRepo,
		mutationRepo,
		dbInstance,
	)

	e.POST("/users", handler.UserRegister(authAggregator))
	e.POST("/users/signin", handler.UserSignin(authAggregator))

	oauthMiddleware := middleware.Oauth(func(c echo.Context, token string) (bool, error) {
		t, err := userTokenRepo.FindByToken(token)
		if err != nil {
			return false, err
		}

		c.Set("current_user", t)
		return true, nil
	})

	e.GET("/wallet", handler.CheckBalance(walletRepo), oauthMiddleware)
	e.GET("/wallet/top-transfer", handler.TopTransfer(mutationRepo), oauthMiddleware)
	e.POST("/transactions/topup", handler.TopUp(trxAggregator), oauthMiddleware)
	e.POST("/transactions/transfer", handler.Transfer(trxAggregator), oauthMiddleware)

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
