package persisters

import (
	"context"
	"database/sql"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (p *Persister) GetContacts(ctx context.Context, namespace string) ([]models.Contact, error) {
	p.log.Debug("Getting contacts", "namespace", namespace)

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
) (int32, error) {
	p.log.Debug("Creating contact", "firstName", firstName, "lastName", lastName, "namespace", namespace)

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
	p.log.Debug("Getting contact", "id", id, "namespace", namespace)

	return p.queries.GetContact(ctx, models.GetContactParams{
		ID:        id,
		Namespace: namespace,
	})
}

func (p *Persister) DeleteContact(ctx context.Context, id int32, namespace string) error {
	p.log.Debug("Deleting contact", "id", id, "namespace", namespace)

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := p.queries.WithTx(tx)

	if err := qtx.DeleteDebtsForContact(ctx, models.DeleteDebtsForContactParams{
		ID:        id,
		Namespace: namespace,
	}); err != nil {
		return err
	}

	if err := qtx.DeleteContact(ctx, models.DeleteContactParams{
		ID:        id,
		Namespace: namespace,
	}); err != nil {
		return err
	}

	return tx.Commit()
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
) error {
	p.log.Debug("Updating contact", "id", id, "firstName", firstName, "lastName", lastName, "namespace", namespace)

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
	p.log.Debug("Counting contacts", "namespace", namespace)

	return p.queries.CountContacts(ctx, namespace)
}
