-- +goose Up
alter table journal_entries
add column rating integer default 2 not null;
alter table journal_entries
add constraint check_raiting check (
        rating >= 1
        and rating <= 3
    );
-- +goose Down
alter table journal_entries drop constraint check_raiting;
alter table journal_entries drop column rating;