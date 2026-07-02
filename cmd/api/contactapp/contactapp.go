package contactapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/jmoiron/sqlx"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/rag"
	"github.com/leogoesger/lead-funnel/internal/web"
	"github.com/pkg/errors"
)

type app struct {
	log       *logger.Logger
	db        *sqlx.DB
	llmClient *llm.Client
	ragClient rag.Client
}

func newApp(log *logger.Logger, db *sqlx.DB, llmClient *llm.Client, ragClient rag.Client) *app {
	return &app{
		log:       log,
		db:        db,
		llmClient: llmClient,
		ragClient: ragClient,
	}
}

type TestRequest struct {
	Question string `json:"question"`
}

func (a *app) test(ctx context.Context, r *http.Request) web.Encoder {
	var req TestRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return web.ErrorResponse{
			Error:   errors.Wrap(err, "invalid request body"),
			Message: "Invalid request body",
			Status:  http.StatusBadRequest,
		}
	}

	if req.Question == "" {
		return web.ErrorResponse{
			Error:  errors.New("question is required"),
			Status: http.StatusBadRequest,
		}
	}

	// always good practice
	defer r.Body.Close()

	results, err := a.ragClient.Query(ctx, req.Question, 3)
	if err != nil {
		return web.ErrorResponse{
			Error:  errors.Wrap(err, "querying RAG system"),
			Status: http.StatusInternalServerError,
		}
	}

	messages := []model.D{
		model.TextMessage("user", req.Question),
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

	ch, err := a.llmClient.ChatStreaming(streamCtx, d)
	if err != nil {
		return web.ErrorResponse{
			Error:  errors.Wrap(err, "chat streaming"),
			Status: http.StatusInternalServerError,
		}
	}

	var reason strings.Builder
	var answer strings.Builder
	for resp := range ch {
		switch resp.Choices[0].FinishReason() {
		case model.FinishReasonError:
			return web.ErrorResponse{
				Error:  errors.New("chat streaming error: " + resp.Choices[0].Delta.Content),
				Status: http.StatusInternalServerError,
			}

		case model.FinishReasonStop:
			a.log.Info(ctx, "reasoning", "reason", reason.String())

		default:
			if resp.Choices[0].Delta.Reasoning != "" {
				reason.WriteString(resp.Choices[0].Delta.Reasoning)
			}
			if resp.Choices[0].Delta.Content != "" {
				answer.WriteString(resp.Choices[0].Delta.Content)
			}
		}
	}
	a.log.Info(ctx, "chat streaming answer", "answer", answer.String())
	return web.SuccessResponse{
		Message: "Chat streaming completed successfully",
		Data:    answer.String(),
	}
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
