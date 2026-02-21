package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bas-x/basex/assert"
	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type RunOptions struct {
	Writer   io.Writer
	Listener net.Listener
	Config   *viper.Viper
}

func Run(ctx context.Context, opts *RunOptions) error {
	assert.NotNil(opts, "run options")
	assert.NotNil(opts.Config, "run config")
	assert.NotNil(opts.Listener, "run listener")
	assert.NotNil(opts.Writer, "run writer")

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := log.Default()
	logger.SetOutput(opts.Writer)
	if opts.Config.GetString("ENVIRONMENT") == "dev" {
		logger.SetLevel(log.DebugLevel)
	}

	deps := initDeps(opts.Config)
	assert.NotNil(deps, "server dependencies")

	server := newServer(
		logger,
		opts.Config,
		deps,
	)
	assert.NotNil(server, "http server")

	go func() {
		log.Info("serving")
		err := server.Serve(opts.Listener)
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("received error from http server", "err", err)
		}
	}()

	<-ctx.Done()

	timeout := time.Second * 10
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("err during shutdown %w", err)
	}

	return nil
}

type ServerDependencies struct {
}

func initDeps(config *viper.Viper) *ServerDependencies {
	assert.NotNil(config, "config")

	return &ServerDependencies{}
}

func newServer(logger *log.Logger, config *viper.Viper, deps *ServerDependencies) *http.Server {
	assert.NotNil(logger, "logger")
	assert.NotNil(config, "config")
	assert.NotNil(deps, "server dependencies")

	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()

	server := &http.Server{
		Handler:           e,
		ReadTimeout:       5 * time.Minute,
		IdleTimeout:       30 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          logger.StandardLog(),
	}

	eLoggerConfig := middleware.DefaultLoggerConfig
	eLoggerConfig.Output = logger.StandardLog().Writer()
	e.Use(middleware.LoggerWithConfig(eLoggerConfig))

	rlimiterConfig := rateLimiterConfig(logger)
	e.Use(middleware.RateLimiterWithConfig(rlimiterConfig))

	corsConfig := corsConfig()
	e.Use(middleware.CORSWithConfig(corsConfig))

	e.Validator = &CustomValidator{validator: validator.New()}

	registerRoutes(e, logger, config, deps)

	assert.NotNil(server, "http server")
	return server
}

func rateLimiterConfig(logger *log.Logger) middleware.RateLimiterConfig {
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(5),
				Burst:     15,
				ExpiresIn: 3 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		DenyHandler: func(ctx echo.Context, identifier string, err error) error {
			logger.Warn("ratelimiting", "identifier", identifier, "err", err)
			return ctx.NoContent(http.StatusTooManyRequests)
		},
	}

	return config
}

func corsConfig() middleware.CORSConfig {
	config := middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderXRequestedWith,
			echo.HeaderAuthorization,
		},
	}

	return config
}

type CustomValidator struct {
	validator *validator.Validate
}

func (v *CustomValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
