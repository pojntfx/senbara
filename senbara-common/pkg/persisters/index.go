package persisters

//go:generate sqlc -f ../../sqlc.yaml generate

import (
	"database/sql"

	"github.com/pojntfx/senbara/senbara-common/pkg/migrations"
	"github.com/pojntfx/senbara/senbara-common/pkg/tables"
	"github.com/pressly/goose/v3"
)

type Persister struct {
	pgaddr  string
	queries *tables.Queries
	db      *sql.DB
}

func NewPersister(pgaddr string) *Persister {
	return &Persister{
		pgaddr: pgaddr,
	}
}

func (p *Persister) Init() error {
	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(p.db, "."); err != nil {
		return err
	}

	p.queries = tables.New(p.db)

	return nil
}
