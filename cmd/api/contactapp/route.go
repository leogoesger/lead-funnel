package contactapp

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/rag"
	"github.com/leogoesger/lead-funnel/internal/web"
)

type Config struct {
	Log       *logger.Logger
	DB        *sqlx.DB
	LLMClient *llm.Client
	RagClient rag.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg *Config) {
	const version = "v1"

	api := newApp(cfg.Log, cfg.DB, cfg.LLMClient, cfg.RagClient)

	app.HandlerFunc(http.MethodPost, version, "/test", api.test)
}
