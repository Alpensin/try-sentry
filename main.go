// Package main - check sentry with go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	zlogsentry "github.com/archdx/zerolog-sentry"
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
	w, err := zlogsentry.New(sentryDSN, zlogsentry.WithEnvironment("dev"), zlogsentry.WithRelease("1.0.0"))
	if err != nil {
		fmt.Printf("zlogsentry initialization failed: %v\n", err)
	}
	defer w.Close()
	multi := zerolog.MultiLevelWriter(os.Stdout, w)
	logger := zerolog.New(multi).With().Timestamp().Logger()
	// Then create your app
	app := echo.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// Once it's done, you can attach the handler as one of your middleware
	app.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
				hub.Scope().SetTag("someRandomTag", "maybeYouNeedIt")
			}
			return next(ctx)
		}
	})
	// Set up routes
	app.GET("/", func(ctx echo.Context) error {
		logger.Error().Err(errors.New("dial timeout")).Str("gang", "bang").Msg("test message")
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
