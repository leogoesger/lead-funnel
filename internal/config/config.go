package config

import (
	"github.com/leogoesger/lead-funnel/internal/db"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/migrate"
	"github.com/leogoesger/lead-funnel/internal/rag"
	"github.com/leogoesger/lead-funnel/internal/web"
)

type ServiceAPI struct {
	Web     *web.Config     `required:"true" split_words:"true"`
	Rag     *rag.Config     `required:"true" split_words:"true"`
	LLM     *llm.Config     `required:"true" split_words:"true"`
	DB      *db.Config      `required:"true" split_words:"true"`
	Migrate *migrate.Config `required:"true" split_words:"true"`
}
