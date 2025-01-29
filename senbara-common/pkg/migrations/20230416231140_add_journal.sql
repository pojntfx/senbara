-- +goose Up
create table journal_entries (
    id serial primary key,
    title text not null,
    date timestamp not null default now(),
    body text not null
);
drop table contacts;
-- +goose Down
drop table journal_entries;
create table contacts (
    id text primary key not null,
    name text not null,
    address text not null
);