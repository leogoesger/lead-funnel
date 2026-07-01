package main

import (
	"context"
	"fmt"
	"os"
	"strings"
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

	db, err := initDB(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to connect to postgres database: %v\n", err)
		return 1
	}
	defer db.Close()

	ragSystem, err := initRagSystem(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create RAG system: %v\n", err)
		return 1
	}

	llmClient, err := initLLM(ctx, &cfg)
	if err != nil {
		log.Error(ctx, "unable to create LLM client: %v\n", err)
		return 1
	}

	if err := testllm(ctx, llmClient, log, ragSystem); err != nil {
		log.Error(ctx, "unable to test LLM client: %v\n", err)
		return 1
	}

	return 0
}

func initDB(ctx context.Context, cfg *config.ServiceAPI) (*sqlx.DB, error) {
	db, err := db.NewPostgres(cfg.DB)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to postgres database")
	}

	migrateService, err := migrate.New(cfg.Migrate, db.DB)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create migration service")
	}

	if err := migrateService.Migrate(ctx); err != nil {
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

	if err := ragSystem.Unload(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to unload RAG system")
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

	var reasoning bool
	for resp := range ch {
		switch resp.Choices[0].FinishReason() {
		case model.FinishReasonError:
			return errors.New("chat streaming error: " + resp.Choices[0].Delta.Content)

		case model.FinishReasonStop:
			return errors.New("finish reason stop")

		default:
			if resp.Choices[0].Delta.Reasoning != "" {
				reasoning = true
				log.Info(ctx, "chat streaming reasoning", "content", fmt.Sprintf("\u001b[91m%s\u001b[0m", resp.Choices[0].Delta.Reasoning))
				continue
			}

			if reasoning {
				reasoning = false
				continue
			}

			log.Info(ctx, "chat streaming response", "content", resp.Choices[0].Delta.Content)
		}
	}
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
