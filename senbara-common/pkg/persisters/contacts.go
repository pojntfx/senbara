package persisters

import (
	"context"
	"database/sql"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (p *Persister) GetContacts(ctx context.Context, namespace string) ([]models.Contact, error) {
	p.log.With("namespace", namespace).Debug("Getting contacts")

	return p.queries.GetContacts(ctx, namespace)
}

func (p *Persister) CreateContact(
	ctx context.Context,
	firstName string,
	lastName string,
	nickname string,
	email string,
	pronouns string,
	namespace string,
) (models.Contact, error) {
	p.log.With("namespace", namespace).Debug("Creating contact", "firstName", firstName, "lastName", lastName)

	return p.queries.CreateContact(ctx, models.CreateContactParams{
		FirstName: firstName,
		LastName:  lastName,
		Nickname:  nickname,
		Email:     email,
		Pronouns:  pronouns,
		Namespace: namespace,
	})
}

func (p *Persister) GetContact(ctx context.Context, id int32, namespace string) (models.Contact, error) {
	p.log.With("namespace", namespace).Debug("Getting contact", "id", id)

	return p.queries.GetContact(ctx, models.GetContactParams{
		ID:        id,
		Namespace: namespace,
	})
}

func (p *Persister) DeleteContact(ctx context.Context, id int32, namespace string) (int32, error) {
	p.log.With("namespace", namespace).Debug("Deleting contact", "id", id)

	tx, err := p.db.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	qtx := p.queries.WithTx(tx)

	if err := qtx.DeleteDebtsForContact(ctx, models.DeleteDebtsForContactParams{
		ID:        id,
		Namespace: namespace,
	}); err != nil {
		return -1, err
	}

	deletedContactID, err := qtx.DeleteContact(ctx, models.DeleteContactParams{
		ID:        id,
		Namespace: namespace,
	})
	if err != nil {
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}

	return deletedContactID, nil
}

func (p *Persister) UpdateContact(
	ctx context.Context,
	id int32,
	firstName,
	lastName,
	nickname,
	email,
	pronouns,
	namespace string,
	birthday *time.Time,
	address,
	notes string,
) (models.Contact, error) {
	p.log.With("namespace", namespace).Debug("Updating contact", "id", id, "firstName", firstName, "lastName", lastName)

	var birthdayDate sql.NullTime
	if birthday != nil {
		birthdayDate = sql.NullTime{
			Time:  *birthday,
			Valid: true,
		}
	}

	return p.queries.UpdateContact(ctx, models.UpdateContactParams{
		ID:        id,
		Namespace: namespace,
		FirstName: firstName,
		LastName:  lastName,
		Nickname:  nickname,
		Email:     email,
		Pronouns:  pronouns,
		Birthday:  birthdayDate,
		Address:   address,
		Notes:     notes,
	})
}

func (p *Persister) CountContacts(ctx context.Context, namespace string) (int64, error) {
	p.log.With("namespace", namespace).Debug("Counting contacts")

	return p.queries.CountContacts(ctx, namespace)
}
