package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Config struct {
	Address  string `required:"true" split_words:"true" desc:"HOST:PORT - if PORT is omitted it will default to 5432"`
	Name     string `required:"true" split_words:"true" desc:"Database name"`
	Username string `required:"true" split_words:"true"`
	Password string `required:"true" split_words:"true"`
}

func NewPostgres(cfg *Config) (*sqlx.DB, error) {
	hp := strings.Split(cfg.Address, ":")
	host := hp[0]

	port := int64(5432)
	if len(hp) > 1 {
		p, err := strconv.ParseInt(hp[1], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "connect to postgres database")
		}
		port = p
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		host,
		port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres database")
	}

	return db, nil
}
