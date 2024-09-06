// Package main - check sentry with go
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Alpensin/try-sentry/pkg/zlog"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func main() {
	err := godotenv.Load()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr})
	if err != nil {
		fmt.Printf("Error loading .env file")
		os.Exit(1)
	}
	sentryDSN := os.Getenv("SENTRY_DSN")
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	sentryLevels := []zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel}
	sentryHook, err := zlog.NewHook(sentryLevels, sentry.ClientOptions{
		Dsn: sentryDSN,
	})
	if err != nil {
		panic("Sentry initialization failed")
	}
	defer sentryHook.Flush(5 * time.Second)
	logger = logger.Hook(sentryHook)
	// Then create your app
	app := echo.New()

	app.Use(middleware.Recover())

	// Set up routes
	app.GET("/", func(ctx echo.Context) error {
		logger.Error().Msg("My error")
		return ctx.String(http.StatusOK, "Hello, World!")
	})
	app.GET("/foo", func(ctx echo.Context) error {
		// sentryecho handler will catch it just fine. Also, because we attached "someRandomTag"
		// in the middleware before, it will be sent through as well
		panic("y tho")
	})
	// And run it
	app.Start(":3000")
}
