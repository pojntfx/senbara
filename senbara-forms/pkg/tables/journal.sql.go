// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: journal.sql

package tables

import (
	"context"
	"time"
)

const countJournalEntries = `-- name: CountJournalEntries :one
select count(*)
from journal_entries
where namespace = $1
`

func (q *Queries) CountJournalEntries(ctx context.Context, namespace string) (int64, error) {
	row := q.db.QueryRowContext(ctx, countJournalEntries, namespace)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createJournalEntry = `-- name: CreateJournalEntry :one
insert into journal_entries (title, body, rating, namespace)
values ($1, $2, $3, $4)
returning id
`

type CreateJournalEntryParams struct {
	Title     string
	Body      string
	Rating    int32
	Namespace string
}

func (q *Queries) CreateJournalEntry(ctx context.Context, arg CreateJournalEntryParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, createJournalEntry,
		arg.Title,
		arg.Body,
		arg.Rating,
		arg.Namespace,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const deleteJournalEntriesForNamespace = `-- name: DeleteJournalEntriesForNamespace :exec
delete from journal_entries
where namespace = $1
`

func (q *Queries) DeleteJournalEntriesForNamespace(ctx context.Context, namespace string) error {
	_, err := q.db.ExecContext(ctx, deleteJournalEntriesForNamespace, namespace)
	return err
}

const deleteJournalEntry = `-- name: DeleteJournalEntry :exec
delete from journal_entries
where id = $1
    and namespace = $2
`

type DeleteJournalEntryParams struct {
	ID        int32
	Namespace string
}

func (q *Queries) DeleteJournalEntry(ctx context.Context, arg DeleteJournalEntryParams) error {
	_, err := q.db.ExecContext(ctx, deleteJournalEntry, arg.ID, arg.Namespace)
	return err
}

const getJournalEntries = `-- name: GetJournalEntries :many
select id, title, date, body, rating, namespace
from journal_entries
where namespace = $1
order by date desc
`

func (q *Queries) GetJournalEntries(ctx context.Context, namespace string) ([]JournalEntry, error) {
	rows, err := q.db.QueryContext(ctx, getJournalEntries, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []JournalEntry
	for rows.Next() {
		var i JournalEntry
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Date,
			&i.Body,
			&i.Rating,
			&i.Namespace,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getJournalEntriesExportForNamespace = `-- name: GetJournalEntriesExportForNamespace :many
select 'journal_entries' as table_name,
    id, title, date, body, rating, namespace
from journal_entries
where namespace = $1
order by date desc
`

type GetJournalEntriesExportForNamespaceRow struct {
	TableName string
	ID        int32
	Title     string
	Date      time.Time
	Body      string
	Rating    int32
	Namespace string
}

func (q *Queries) GetJournalEntriesExportForNamespace(ctx context.Context, namespace string) ([]GetJournalEntriesExportForNamespaceRow, error) {
	rows, err := q.db.QueryContext(ctx, getJournalEntriesExportForNamespace, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetJournalEntriesExportForNamespaceRow
	for rows.Next() {
		var i GetJournalEntriesExportForNamespaceRow
		if err := rows.Scan(
			&i.TableName,
			&i.ID,
			&i.Title,
			&i.Date,
			&i.Body,
			&i.Rating,
			&i.Namespace,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getJournalEntry = `-- name: GetJournalEntry :one
select id, title, date, body, rating, namespace
from journal_entries
where id = $1
    and namespace = $2
`

type GetJournalEntryParams struct {
	ID        int32
	Namespace string
}

func (q *Queries) GetJournalEntry(ctx context.Context, arg GetJournalEntryParams) (JournalEntry, error) {
	row := q.db.QueryRowContext(ctx, getJournalEntry, arg.ID, arg.Namespace)
	var i JournalEntry
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Date,
		&i.Body,
		&i.Rating,
		&i.Namespace,
	)
	return i, err
}

const updateJournalEntry = `-- name: UpdateJournalEntry :exec
update journal_entries
set title = $3,
    body = $4,
    rating = $5
where id = $1
    and namespace = $2
`

type UpdateJournalEntryParams struct {
	ID        int32
	Namespace string
	Title     string
	Body      string
	Rating    int32
}

func (q *Queries) UpdateJournalEntry(ctx context.Context, arg UpdateJournalEntryParams) error {
	_, err := q.db.ExecContext(ctx, updateJournalEntry,
		arg.ID,
		arg.Namespace,
		arg.Title,
		arg.Body,
		arg.Rating,
	)
	return err
}
