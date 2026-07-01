package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/ardanlabs/kronk/sdk/tools/models"
)

type Client struct {
	*kronk.Kronk
}

type Config struct {
	Model string `default:"unsloth/Qwen3-0.6B-Q8_0" split_words:"true"`
}

func New(ctx context.Context, c *Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	mdl, err := models.New()
	if err != nil {
		return nil, fmt.Errorf("unable to create model: %w", err)
	}

	mp, err := mdl.Download(ctx, kronk.FmtLogger, c.Model)
	if err != nil {
		return nil, fmt.Errorf("unable to download model: %w", err)
	}

	if err := kronk.Init(); err != nil {
		return nil, fmt.Errorf("unable to initialize kronk: %w", err)
	}

	krn, err := kronk.New(
		model.WithModelFiles(mp.ModelFiles),
		model.WithAutoTune(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create kronk: %w", err)
	}

	return &Client{Kronk: krn}, nil
}
