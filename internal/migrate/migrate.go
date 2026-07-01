package migrate

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/leogoesger/lead-funnel/internal/logger"
	"github.com/pkg/errors"
)

type Config struct {
	Name                    string `default:"contentmods" split_words:"true"`
	MigrationFile           string `default:"file:///migrations" split_words:"true"`
	SeedFile                string `split_words:"true"`
	DBReadyMaxCheckAttempts int    `default:"10" split_words:"true"`
}

type migrationService struct {
	cfg                     Config
	db                      *sql.DB
	migrate                 *migrate.Migrate
	dbReadyMaxCheckAttempts int
	log                     *logger.Logger
}

func New(cfg *Config, db *sql.DB) (*migrationService, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "get driver")
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.MigrationFile,
		cfg.Name,
		driver,
	)

	if err != nil {
		return nil, errors.Wrap(err, "create migrate instance")
	}

	return &migrationService{
		cfg:                     *cfg,
		db:                      db,
		migrate:                 m,
		dbReadyMaxCheckAttempts: cfg.DBReadyMaxCheckAttempts,
	}, nil
}

func (s *migrationService) Migrate(ctx context.Context) error {
	err := s.readyDB(ctx)
	if err != nil {
		return errors.Wrap(err, "ready db")
	}

	if err := s.migrateDB(ctx); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	return nil
}

func (s *migrationService) readyDB(ctx context.Context) error {
	var pingError error
	for attempts := 1; attempts <= s.dbReadyMaxCheckAttempts; attempts++ {
		pingError = s.db.Ping()
		if pingError == nil {
			break
		}
		s.log.Info(ctx, "waiting for database to become ready")
		time.Sleep(time.Duration(attempts) * time.Second)
	}

	return pingError
}

func (s *migrationService) migrateDB(ctx context.Context) error {
	err := s.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "apply migrations")
	}

	ver, dirty, err := s.migrate.Version()
	s.log.Info(ctx, "database version", "version", ver, "dirty", dirty, "error", err)

	if s.cfg.SeedFile != "" {
		seedScript, err := os.ReadFile(s.cfg.SeedFile)
		if err != nil {
			return errors.Wrap(err, "read seed file")
		}

		if _, err := s.db.Exec(string(seedScript)); err != nil {
			return errors.Wrap(err, "run seed script")
		}
	}

	return nil
}
