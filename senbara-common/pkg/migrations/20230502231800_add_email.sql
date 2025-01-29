-- +goose Up
alter table journal_entries
add column namespace text default '' not null;
-- +goose Down
alter table journal_entries drop column namespace;