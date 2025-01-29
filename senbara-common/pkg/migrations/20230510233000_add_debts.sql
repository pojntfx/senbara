-- +goose Up
create table debts (
    id serial primary key,
    amount float not null,
    currency text not null,
    contact_id integer not null,
    foreign key (contact_id) references contacts (id)
);
-- +goose Down
drop table debts;