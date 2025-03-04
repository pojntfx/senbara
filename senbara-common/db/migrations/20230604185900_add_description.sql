-- +goose Up
alter table debts
add column description text default '' not null;
-- +goose Down
alter table debts drop column description;