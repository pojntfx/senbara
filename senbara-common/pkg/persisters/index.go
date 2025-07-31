package persisters

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc -f ../../sqlc.yaml generate

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pojntfx/senbara/senbara-common/db/migrations"
	"github.com/pojntfx/senbara/senbara-common/internal/tables"
	"github.com/pojntfx/senbara/senbara-common/pkg/models"
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
	p.log.Info("Connecting to database")

	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	goose.SetLogger(slog.NewLogLogger(p.log.Handler(), slog.LevelDebug))
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

func (p *Persister) CountContactsAndJournalEntries(ctx context.Context, namespace string) (models.ContactsAndJournalEntriesCount, error) {
	p.log.With("namespace", namespace).Debug("Counting contacts and journal entries")

	return p.queries.CountContactsAndJournalEntries(ctx, namespace)
}

func (p *Persister) CountAllContactsAndJournalEntries(ctx context.Context) (models.ContactsAndJournalEntriesCount, error) {
	p.log.Debug("Counting all contacts and journal entries")

	allContactsAndJournalEntriesCount, err := p.queries.CountAllContactsAndJournalEntries(ctx)
	if err != nil {
		return tables.CountContactsAndJournalEntriesRow{}, err
	}

	return models.ContactsAndJournalEntriesCount(allContactsAndJournalEntriesCount), nil
}
