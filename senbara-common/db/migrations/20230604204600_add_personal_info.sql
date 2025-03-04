-- +goose Up
alter table contacts
add column birthday date,
    add column address text default '' not null,
    add column notes text default '' not null;
-- +goose Down
alter table contacts drop column birthday,
    drop column address,
    drop column notes;