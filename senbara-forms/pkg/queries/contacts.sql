-- name: GetContacts :many
select *
from contacts
where namespace = $1
order by first_name desc;
-- name: CreateContact :one
insert into contacts (
        first_name,
        last_name,
        nickname,
        email,
        pronouns,
        namespace
    )
values ($1, $2, $3, $4, $5, $6)
returning id;
-- name: DeleteContact :exec
delete from contacts
where id = $1
    and namespace = $2;
-- name: GetContact :one
select *
from contacts
where id = $1
    and namespace = $2;
-- name: UpdateContact :exec
update contacts
set first_name = $3,
    last_name = $4,
    nickname = $5,
    email = $6,
    pronouns = $7,
    birthday = $8,
    address = $9,
    notes = $10
where id = $1
    and namespace = $2;
-- name: DeleteContactsForNamespace :exec
delete from contacts
where namespace = $1;
-- name: GetContactsExportForNamespace :many
select 'contacts' as table_name,
    *
from contacts
where namespace = $1
order by first_name desc;
-- name: CountContacts :one
select count(*)
from contacts
where namespace = $1;