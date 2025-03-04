-- +goose Up
create table contacts (
    id serial primary key,
    first_name text not null,
    last_name text not null,
    nickname text not null,
    email text not null,
    pronouns text not null,
    namespace text not null
);
-- +goose Down
drop table contacts;