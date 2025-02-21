-- name: GetJournalEntries :many
select *
from journal_entries
where namespace = $1
order by date desc;
-- name: GetJournalEntry :one
select *
from journal_entries
where id = $1
    and namespace = $2;
-- name: CreateJournalEntry :one
insert into journal_entries (title, body, rating, namespace)
values ($1, $2, $3, $4)
returning *;
-- name: DeleteJournalEntry :one
delete from journal_entries
where id = $1
    and namespace = $2
returning id;
-- name: UpdateJournalEntry :one
update journal_entries
set title = $3,
    body = $4,
    rating = $5
where id = $1
    and namespace = $2
returning *;
-- name: DeleteJournalEntriesForNamespace :many
delete from journal_entries
where namespace = $1
returning id;
-- name: GetJournalEntriesExportForNamespace :many
select 'journal_entries' as table_name,
    *
from journal_entries
where namespace = $1
order by date desc;
-- name: CountJournalEntries :one
select count(*)
from journal_entries
where namespace = $1;