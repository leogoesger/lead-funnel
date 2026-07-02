package migrate

import (
	"database/sql"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
)

type Config struct {
	Name                    string `default:"lead_funnel" split_words:"true"`
	MigrationFile           string `default:"file://./migrations" split_words:"true"`
	SeedFile                string `split_words:"true"`
	DBReadyMaxCheckAttempts int    `default:"10" split_words:"true"`
}

type migrationService struct {
	cfg                     Config
	db                      *sql.DB
	migrate                 *migrate.Migrate
	dbReadyMaxCheckAttempts int
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

func (s *migrationService) Migrate() error {
	err := s.readyDB()
	if err != nil {
		return errors.Wrap(err, "ready db")
	}

	if err := s.migrateDB(); err != nil {
		return errors.Wrap(err, "migrate db")
	}

	return nil
}

func (s *migrationService) readyDB() error {
	var pingError error
	for attempts := 1; attempts <= s.dbReadyMaxCheckAttempts; attempts++ {
		pingError = s.db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * time.Second)
	}

	return pingError
}

func (s *migrationService) migrateDB() error {
	err := s.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "apply migrations")
	}

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
