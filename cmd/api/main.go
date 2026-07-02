package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
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

	ragSystem, err := initRagSystem(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create RAG system", "err", err)
		return 1
	}

	llmClient, err := initLLM(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create LLM client", "err", err)
		return 1
	}

	if err := testllm(ctx, llmClient, log, ragSystem); err != nil {
		log.Error(ctx, "unable to test LLM client", "err", err)
		return 1
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := web.MuxConfig{
		Log: log,
		DB:  db,
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

func initRagSystem(ctx context.Context, cfg *config.ServiceAPI) (rag.RAG, error) {
	embedder, err := rag.NewKronkEmbedder(ctx, cfg.Rag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create RAG embedder")
	}

	vectorStore := rag.NewInMemoryVectorStore(cfg.Rag)

	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg.Rag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create RAG system")
	}

	if err := ragSystem.LoadDocuments(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to load documents")
	}

	return ragSystem, nil
}

func initLLM(ctx context.Context, cfg *config.ServiceAPI) (*llm.Client, error) {
	llmClient, err := llm.New(ctx, cfg.LLM)
	if err != nil {
		return nil, errors.Wrap(err, "create LLM client")
	}

	return llmClient, nil
}

func testllm(ctx context.Context, llmClient *llm.Client, log *logger.Logger, ragSystem rag.RAG) error {
	question := "What is your address?"
	results, err := ragSystem.Query(ctx, question, 3)
	if err != nil {
		return errors.Wrap(err, "query RAG system")
	}

	messages := []model.D{
		model.TextMessage("user", question),
	}

	d := model.D{
		"messages":    addContextPrompt(results, messages),
		"max_tokens":  2048,
		"temperature": 0.7,
		"top_p":       0.9,
		"top_k":       40,
	}

	streamCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ch, err := llmClient.ChatStreaming(streamCtx, d)
	if err != nil {
		return errors.Wrap(err, "chat streaming")
	}

	var reason strings.Builder
	var answer strings.Builder
	for resp := range ch {
		switch resp.Choices[0].FinishReason() {
		case model.FinishReasonError:
			return errors.New("chat streaming error: " + resp.Choices[0].Delta.Content)

		case model.FinishReasonStop:
			log.Info(ctx, "reasoning", "reason", reason.String())

		default:
			if resp.Choices[0].Delta.Reasoning != "" {
				reason.WriteString(resp.Choices[0].Delta.Reasoning)
			}

			if resp.Choices[0].Delta.Content != "" {
				answer.WriteString(resp.Choices[0].Delta.Content)
			}
		}
	}
	log.Info(ctx, "chat streaming answer", "answer", answer.String())
	return nil
}

func addContextPrompt(documents []rag.QueryResult, messages []model.D) []model.D {
	const prompt = `
		- Use the following Context to answer the user's question.
		- If you don't know the answer, say that you don't know.
		- Responses should be properly formatted to be easily read.
		- Share code if code is presented in the context.
		- Do not include any additional information not present in the context.

		Context:
		
		%s

		Question: %s
		`

	var count int
	var content strings.Builder
	for _, doc := range documents {
		content.WriteString(fmt.Sprintf("%s\n\n", doc.Chunk.Content))
		count++
		if count == 2 {
			break
		}
	}

	lastUserInput := messages[len(messages)-1]["content"].(string)
	finalPrompt := fmt.Sprintf(prompt, content.String(), lastUserInput)

	messages = append(messages, model.TextMessage("user", finalPrompt))

	return messages
}
