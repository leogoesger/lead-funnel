package healthapp

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/leogoesger/lead-funnel/internal/web"
)

type app struct {
	log *logger.Logger
	db  *sqlx.DB
}

func newApp(log *logger.Logger, db *sqlx.DB) *app {
	return &app{
		log: log,
		db:  db,
	}
}

func (a *app) health(ctx context.Context, r *http.Request) web.Encoder {
	return web.SuccessResponse{
		Message: "Service is healthy",
		Data: map[string]string{
			"status": "ok",
		},
	}
}
