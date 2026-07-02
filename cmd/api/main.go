package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/leogoesger/lead-funnel/internal/config"
	"github.com/leogoesger/lead-funnel/internal/db"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/migrate"
	"github.com/leogoesger/lead-funnel/internal/rag"
	"github.com/leogoesger/lead-funnel/internal/web"
	"github.com/pkg/errors"
)

func main() {
	ctx := context.Background()
	exitCode := run(ctx)
	os.Exit(exitCode)
}

func GetTraceID(ctx context.Context) string {
	return "00000000-0000-0000-0000-000000000000"
}

func run(ctx context.Context) int {
	log := logger.New(os.Stdout, logger.LevelInfo, "SERVER", GetTraceID)
	var cfg config.ServiceAPI
	if err := envconfig.Process("", &cfg); err != nil {
		_ = envconfig.Usage("", &cfg)
		log.Error(ctx, "load config from environment: %v\n", err)
		return 1
	}

	db, err := initDB(&cfg)
	if err != nil {
		log.Error(ctx, "unable to connect to postgres database", "err", err)
		return 1
	}
	defer db.Close()

	ragClient, err := initRagClient(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create RAG system", "err", err)
		return 1
	}

	llmClient, err := initLLM(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create LLM client", "err", err)
		return 1
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := web.MuxConfig{
		Log:       log,
		DB:        db,
		LLMClient: llmClient,
		RagClient: ragClient,
	}

	webAPI := web.WebAPI(cfgMux,
		Routes(),
		web.WithCORS(cfg.Web.CORSAllowedOrigins),
	)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      webAPI,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		log.Error(ctx, "server error", "err", err)
		return 1

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return 1
		}

		log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)
	}

	return 0
}

func initDB(cfg *config.ServiceAPI) (*sqlx.DB, error) {
	db, err := db.NewPostgres(cfg.DB)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to postgres database")
	}

	migrateService, err := migrate.New(cfg.Migrate, db.DB)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create migration service")
	}

	if err := migrateService.Migrate(); err != nil {
		return nil, errors.Wrap(err, "unable to migrate database")
	}

	return db, nil
}

func initRagClient(ctx context.Context, cfg *config.ServiceAPI) (rag.Client, error) {
	embedder, err := rag.NewKronkEmbedder(ctx, cfg.Rag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create RAG embedder")
	}

	vectorStore := rag.NewInMemoryVectorStore(cfg.Rag)

	ragClient, err := rag.NewRAG(vectorStore, embedder, cfg.Rag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create RAG system")
	}

	if err := ragClient.LoadDocuments(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to load documents")
	}

	return ragClient, nil
}

func initLLM(ctx context.Context, cfg *config.ServiceAPI) (*llm.Client, error) {
	llmClient, err := llm.New(ctx, cfg.LLM)
	if err != nil {
		return nil, errors.Wrap(err, "create LLM client")
	}

	return llmClient, nil
}
