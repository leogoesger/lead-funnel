package web

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/rag"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin []string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origins []string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origins
	}
}

type MuxConfig struct {
	Log       *logger.Logger
	DB        *sqlx.DB
	LLMClient *llm.Client
	RagClient rag.Client
}

type RouteAdder interface {
	Add(app *App, cfg MuxConfig)
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg MuxConfig, routeAdder RouteAdder, options ...func(opts *Options)) http.Handler {
	app := NewApp(
		cfg.Log.Info,
		MidTracer(),
		MidLogger(cfg.Log),
		MidErrors(cfg.Log),
	)

	var opts Options
	for _, option := range options {
		option(&opts)
	}

	if len(opts.corsOrigin) > 0 {
		app.EnableCORS(opts.corsOrigin)
	}

	routeAdder.Add(app, cfg)

	app.NotFoundHandler()

	return app
}
