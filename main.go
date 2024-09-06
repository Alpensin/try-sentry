// Package main - check sentry with go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Alpensin/try-sentry/pkg/zlog"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr})
	sentryLevels := []zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel}
	sentryDSN := os.Getenv("SENTRY_DSN")
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

	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// Once it's done, you can attach the handler as one of your middleware
	app.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))
	// Set up routes
	app.GET("/", func(ctx echo.Context) error {
		err := errors.New("seems we have an error here")
		logger.Error().Err(err).Msg("My error")
		return ctx.String(http.StatusOK, "Hello, World!")
	})
	app.GET("/foo", func(ctx echo.Context) error {
		// sentryecho handler will catch it just fine. Also, because we attached "someRandomTag"
		// in the middleware before, it will be sent through as well
		panic("y tho")
	})
	// And run it
	app.Logger.Fatal(app.Start(":3000"))
}
