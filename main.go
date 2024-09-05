// Package main - check sentry with go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file")
		os.Exit(1)
	}
	sentryDSN := os.Getenv("SENTRY_DSN")
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for tracing.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// Then create your app
	app := echo.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// Once it's done, you can attach the handler as one of your middleware
	app.Use(sentryecho.New(sentryecho.Options{}))

	// Set up routes
	app.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	// And run it
	app.Logger.Fatal(app.Start(":3000"))
}
