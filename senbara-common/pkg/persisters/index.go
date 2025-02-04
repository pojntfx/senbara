package persisters

//go:generate sqlc -f ../../sqlc.yaml generate

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pojntfx/senbara/senbara-common/pkg/migrations"
	"github.com/pojntfx/senbara/senbara-common/pkg/tables"
	"github.com/pressly/goose/v3"
)

type Persister struct {
	log     *slog.Logger
	pgaddr  string
	queries *tables.Queries
	db      *sql.DB
}

func NewPersister(log *slog.Logger, pgaddr string) *Persister {
	return &Persister{
		log:    log,
		pgaddr: pgaddr,
	}
}

func (p *Persister) Init(ctx context.Context) error {
	if p.log.Enabled(ctx, slog.LevelDebug) {
		p.log.Debug("Connecting to database", "addr", p.pgaddr)
	} else {
		p.log.Info("Connecting to database")
	}

	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	if !p.log.Enabled(ctx, slog.LevelDebug) {
		goose.SetLogger(goose.NopLogger())
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	p.log.Info("Running migrations")

	if err := goose.Up(p.db, "."); err != nil {
		return err
	}

	p.queries = tables.New(p.db)

	return nil
}
