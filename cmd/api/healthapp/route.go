package healthapp

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *logger.Logger
	DB  *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg *Config) {
	const version = "v1"

	api := newApp(cfg.Log, cfg.DB)

	app.HandlerFunc(http.MethodGet, version, "/health", api.health)
}
