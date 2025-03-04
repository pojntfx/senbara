-- +goose Up
create table activities (
    id serial primary key,
    name text not null,
    date timestamp not null default now(),
    contact_id integer not null,
    description text not null,
    foreign key (contact_id) references contacts (id)
);
-- +goose Down
drop table activities;