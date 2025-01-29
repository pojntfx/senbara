-- name: CreateActivity :one
with contact as (
    select id
    from contacts
    where contacts.id = $1
        and namespace = $2
),
insertion as (
    insert into activities (name, date, description, contact_id)
    select $3,
        $4,
        $5,
        $1
    from contact
    where exists (
            select 1
            from contact
        )
    returning activities.id
)
select id
from insertion;
-- name: GetActivities :many
select activities.id,
    activities.name,
    activities.date,
    activities.description
from contacts
    right join activities on activities.contact_id = contacts.id
where contacts.id = $1
    and contacts.namespace = $2;
-- name: GetActivity :one
select *
from contacts
where contacts.id = $1
    and contacts.namespace = $2;
-- name: DeleteActivity :exec
delete from activities using contacts
where activities.id = $3
    and activities.contact_id = contacts.id
    and contacts.id = $1
    and contacts.namespace = $2;
-- name: DeleteActivitesForContact :exec
delete from activities using contacts
where activities.contact_id = contacts.id
    and contacts.id = $1
    and contacts.namespace = $2;
-- name: GetActivityAndContact :one
select activities.id as activity_id,
    activities.name,
    activities.date,
    activities.description,
    contacts.id as contact_id,
    contacts.first_name,
    contacts.last_name
from contacts
    inner join activities on activities.contact_id = contacts.id
where contacts.id = $1
    and contacts.namespace = $2
    and activities.id = $3;
-- name: UpdateActivity :exec
update activities
set name = $4,
    date = $5,
    description = $6
from contacts
where contacts.id = $1
    and contacts.namespace = $2
    and activities.id = $3
    and activities.contact_id = contacts.id;
-- name: GetActivitiesExportForNamespace :many
select 'activites' as table_name,
    activities.id,
    activities.name,
    activities.date,
    activities.description,
    contacts.id as contact_id
from contacts
    right join activities on activities.contact_id = contacts.id
where contacts.namespace = $1;
-- name: DeleteActivitiesForNamespace :exec
delete from activities using contacts
where activities.contact_id = contacts.id
    and contacts.namespace = $1;