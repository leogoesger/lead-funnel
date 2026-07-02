package main

import (
	"github.com/leogoesger/lead-funnel/cmd/api/healthapp"
	"github.com/leogoesger/lead-funnel/internal/web"
)

func Routes() all {
	return all{}
}

type all struct{}

// Add implements the RouterAdder interface.
func (all) Add(app *web.App, cfg web.MuxConfig) {
	healthapp.Routes(app, &healthapp.Config{
		Log: cfg.Log,
		DB:  cfg.DB,
	})
}
