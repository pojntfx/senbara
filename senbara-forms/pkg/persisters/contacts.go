package persisters

import (
	"context"
	"database/sql"
	"time"

	"github.com/pojntfx/senbara/senbara-forms/pkg/models"
)

func (p *Persister) GetContacts(ctx context.Context, namespace string) ([]models.Contact, error) {
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
	return p.queries.GetContact(ctx, models.GetContactParams{
		ID:        id,
		Namespace: namespace,
	})
}

func (p *Persister) DeleteContact(ctx context.Context, id int32, namespace string) error {
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
	return p.queries.CountContacts(ctx, namespace)
}
