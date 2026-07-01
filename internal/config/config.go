package config

import (
	"github.com/leogoesger/lead-funnel/internal/db"
	"github.com/leogoesger/lead-funnel/internal/llm"
	"github.com/leogoesger/lead-funnel/internal/migrate"
	"github.com/leogoesger/lead-funnel/internal/rag"
)

type ServiceAPI struct {
	Rag     *rag.Config     `required:"true" split_words:"true"`
	LLM     *llm.Config     `required:"true" split_words:"true"`
	DB      *db.Config      `required:"true" split_words:"true"`
	Migrate *migrate.Config `required:"true" split_words:"true"`
}
