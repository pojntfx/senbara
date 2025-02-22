// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: activities.sql

package tables

import (
	"context"
	"database/sql"
	"time"
)

const createActivity = `-- name: CreateActivity :one
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
    returning activities.id,
        activities.name,
        activities.date,
        activities.description
)
select id,
    name,
    date,
    description
from insertion
`

type CreateActivityParams struct {
	ID          int32
	Namespace   string
	Name        string
	Date        time.Time
	Description string
}

type CreateActivityRow struct {
	ID          int32
	Name        string
	Date        time.Time
	Description string
}

func (q *Queries) CreateActivity(ctx context.Context, arg CreateActivityParams) (CreateActivityRow, error) {
	row := q.db.QueryRowContext(ctx, createActivity,
		arg.ID,
		arg.Namespace,
		arg.Name,
		arg.Date,
		arg.Description,
	)
	var i CreateActivityRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Date,
		&i.Description,
	)
	return i, err
}

const deleteActivitesForContact = `-- name: DeleteActivitesForContact :exec
delete from activities using contacts
where activities.contact_id = contacts.id
    and contacts.id = $1
    and contacts.namespace = $2
`

type DeleteActivitesForContactParams struct {
	ID        int32
	Namespace string
}

func (q *Queries) DeleteActivitesForContact(ctx context.Context, arg DeleteActivitesForContactParams) error {
	_, err := q.db.ExecContext(ctx, deleteActivitesForContact, arg.ID, arg.Namespace)
	return err
}

const deleteActivitiesForNamespace = `-- name: DeleteActivitiesForNamespace :many
delete from activities using contacts
where activities.contact_id = contacts.id
    and contacts.namespace = $1
returning activities.id
`

func (q *Queries) DeleteActivitiesForNamespace(ctx context.Context, namespace string) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, deleteActivitiesForNamespace, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteActivity = `-- name: DeleteActivity :one
delete from activities using contacts
where activities.id = $1
    and activities.contact_id = contacts.id
    and contacts.namespace = $2
returning activities.id
`

type DeleteActivityParams struct {
	ID        int32
	Namespace string
}

func (q *Queries) DeleteActivity(ctx context.Context, arg DeleteActivityParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, deleteActivity, arg.ID, arg.Namespace)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getActivities = `-- name: GetActivities :many
select activities.id,
    activities.name,
    activities.date,
    activities.description
from contacts
    right join activities on activities.contact_id = contacts.id
where contacts.id = $1
    and contacts.namespace = $2
`

type GetActivitiesParams struct {
	ID        int32
	Namespace string
}

type GetActivitiesRow struct {
	ID          int32
	Name        string
	Date        time.Time
	Description string
}

func (q *Queries) GetActivities(ctx context.Context, arg GetActivitiesParams) ([]GetActivitiesRow, error) {
	rows, err := q.db.QueryContext(ctx, getActivities, arg.ID, arg.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetActivitiesRow
	for rows.Next() {
		var i GetActivitiesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Date,
			&i.Description,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getActivitiesExportForNamespace = `-- name: GetActivitiesExportForNamespace :many
select 'activites' as table_name,
    activities.id,
    activities.name,
    activities.date,
    activities.description,
    contacts.id as contact_id
from contacts
    right join activities on activities.contact_id = contacts.id
where contacts.namespace = $1
`

type GetActivitiesExportForNamespaceRow struct {
	TableName   string
	ID          int32
	Name        string
	Date        time.Time
	Description string
	ContactID   sql.NullInt32
}

func (q *Queries) GetActivitiesExportForNamespace(ctx context.Context, namespace string) ([]GetActivitiesExportForNamespaceRow, error) {
	rows, err := q.db.QueryContext(ctx, getActivitiesExportForNamespace, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetActivitiesExportForNamespaceRow
	for rows.Next() {
		var i GetActivitiesExportForNamespaceRow
		if err := rows.Scan(
			&i.TableName,
			&i.ID,
			&i.Name,
			&i.Date,
			&i.Description,
			&i.ContactID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getActivityAndContact = `-- name: GetActivityAndContact :one
select activities.id as activity_id,
    activities.name,
    activities.date,
    activities.description,
    contacts.id as contact_id,
    contacts.first_name,
    contacts.last_name
from contacts
    inner join activities on activities.contact_id = contacts.id
where activities.id = $1
    and contacts.namespace = $2
`

type GetActivityAndContactParams struct {
	ID        int32
	Namespace string
}

type GetActivityAndContactRow struct {
	ActivityID  int32
	Name        string
	Date        time.Time
	Description string
	ContactID   int32
	FirstName   string
	LastName    string
}

func (q *Queries) GetActivityAndContact(ctx context.Context, arg GetActivityAndContactParams) (GetActivityAndContactRow, error) {
	row := q.db.QueryRowContext(ctx, getActivityAndContact, arg.ID, arg.Namespace)
	var i GetActivityAndContactRow
	err := row.Scan(
		&i.ActivityID,
		&i.Name,
		&i.Date,
		&i.Description,
		&i.ContactID,
		&i.FirstName,
		&i.LastName,
	)
	return i, err
}

const updateActivity = `-- name: UpdateActivity :one
update activities
set name = $3,
    date = $4,
    description = $5
from contacts
where activities.id = $1
    and contacts.namespace = $2
    and activities.contact_id = contacts.id
returning activities.id,
    activities.name,
    activities.date,
    activities.description
`

type UpdateActivityParams struct {
	ID          int32
	Namespace   string
	Name        string
	Date        time.Time
	Description string
}

type UpdateActivityRow struct {
	ID          int32
	Name        string
	Date        time.Time
	Description string
}

func (q *Queries) UpdateActivity(ctx context.Context, arg UpdateActivityParams) (UpdateActivityRow, error) {
	row := q.db.QueryRowContext(ctx, updateActivity,
		arg.ID,
		arg.Namespace,
		arg.Name,
		arg.Date,
		arg.Description,
	)
	var i UpdateActivityRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Date,
		&i.Description,
	)
	return i, err
}
