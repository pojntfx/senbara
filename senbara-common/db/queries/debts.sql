-- name: CreateDebt :one
with contact as (
    select id
    from contacts
    where contacts.id = $1
        and namespace = $2
),
insertion as (
    insert into debts (amount, currency, description, contact_id)
    select $3,
        $4,
        $5,
        $1
    from contact
    where exists (
            select 1
            from contact
        )
    returning debts.id,
        debts.amount,
        debts.currency,
        debts.description
)
select id,
    amount,
    currency,
    description
from insertion;

-- name: GetDebts :many
select debts.id,
    debts.amount,
    debts.currency,
    debts.description
from contacts
    right join debts on debts.contact_id = contacts.id
where contacts.id = $1
    and contacts.namespace = $2;

-- name: SettleDebt :one
delete from debts using contacts
where debts.id = $1
    and debts.contact_id = contacts.id
    and contacts.namespace = $2
returning debts.id;

-- name: DeleteDebtsForContact :exec
delete from debts using contacts
where debts.contact_id = contacts.id
    and contacts.id = $1
    and contacts.namespace = $2;

-- name: GetDebtAndContact :one
select debts.id as debt_id,
    debts.amount,
    debts.currency,
    debts.description,
    contacts.id as contact_id,
    contacts.first_name,
    contacts.last_name
from contacts
    inner join debts on debts.contact_id = contacts.id
where debts.id = $1
    and contacts.namespace = $2;

-- name: UpdateDebt :one
update debts
set amount = $3,
    currency = $4,
    description = $5
from contacts
where debts.id = $1
    and contacts.namespace = $2
    and debts.contact_id = contacts.id
returning debts.id,
    debts.amount,
    debts.currency,
    debts.description;

-- name: GetDebtsExportForNamespace :many
select 'debts' as table_name,
    debts.id,
    debts.amount,
    debts.currency,
    debts.description,
    contacts.id as contact_id
from contacts
    right join debts on debts.contact_id = contacts.id
where contacts.namespace = $1;

-- name: DeleteDebtsForNamespace :many
delete from debts using contacts
where debts.contact_id = contacts.id
    and contacts.namespace = $1
returning debts.id;