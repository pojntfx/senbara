package persisters

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

var (
	ErrContactDoesNotExist = errors.New("contact does not exist")
)

func (p *Persister) GetUserData(
	ctx context.Context,

	namespace string,

	onJournalEntry func(journalEntry models.ExportedJournalEntry) error,
	onContact func(contact models.ExportedContact) error,
	onDebt func(debt models.ExportedDebt) error,
	onActivity func(activity models.ExportedActivity) error,
) error {
	p.log.With("namespace", namespace).Debug("Getting user data")

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := p.queries.WithTx(tx)

	journalEntries, err := qtx.GetJournalEntriesExportForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	for _, journalEntry := range journalEntries {
		p.log.With("namespace", namespace).Debug("Fetched journal entry", "journalEntryID", journalEntry.ID, "title", journalEntry.Title, "date", journalEntry.Date, "rating", journalEntry.Rating)

		if err := onJournalEntry(models.ExportedJournalEntry{
			ID:        journalEntry.ID,
			Title:     journalEntry.Title,
			Date:      journalEntry.Date,
			Body:      journalEntry.Body,
			Rating:    journalEntry.Rating,
			Namespace: journalEntry.Namespace,
		}); err != nil {
			return err
		}
	}

	contacts, err := qtx.GetContactsExportForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	for _, contact := range contacts {
		p.log.With("namespace", namespace).Debug("Fetched contact", "contactID", contact.ID, "firstName", contact.FirstName, "lastName", contact.LastName, "email", contact.Email)

		if err := onContact(models.ExportedContact{
			ID:        contact.ID,
			FirstName: contact.FirstName,
			LastName:  contact.LastName,
			Nickname:  contact.Nickname,
			Email:     contact.Email,
			Pronouns:  contact.Pronouns,
			Namespace: contact.Namespace,
			Birthday:  contact.Birthday,
			Address:   contact.Address,
			Notes:     contact.Notes,
		}); err != nil {
			return err
		}
	}

	debts, err := qtx.GetDebtsExportForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	for _, debt := range debts {
		p.log.With("namespace", namespace).Debug("Fetched debt", "debtID", debt.ID, "amount", debt.Amount, "currency", debt.Currency, "contactID", debt.ContactID)

		if err := onDebt(models.ExportedDebt{
			ID:          debt.ID,
			Amount:      debt.Amount,
			Currency:    debt.Currency,
			Description: debt.Description,
			ContactID:   debt.ContactID,
		}); err != nil {
			return err
		}
	}

	activities, err := qtx.GetActivitiesExportForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	for _, activity := range activities {
		p.log.With("namespace", namespace).Debug("Fetched activity", "activityID", activity.ID, "name", activity.Name, "date", activity.Date, "contactID", activity.ContactID)

		if err := onActivity(models.ExportedActivity{
			ID:          activity.ID,
			Name:        activity.Name,
			Date:        activity.Date,
			Description: activity.Description,
			ContactID:   activity.ContactID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *Persister) DeleteUserData(ctx context.Context, namespace string) error {
	log := p.log.With("namespace", namespace)

	log.Debug("Deleting user data")

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := p.queries.WithTx(tx)

	activityIDs, err := qtx.DeleteActivitiesForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	log.With("len", len(activityIDs)).Debug("Deleted activities")

	debtIDs, err := qtx.DeleteDebtsForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	log.With("len", len(debtIDs)).Debug("Deleted debts")

	contactIDs, err := qtx.DeleteContactsForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	log.With("len", len(contactIDs)).Debug("Deleted debts")

	journalEntryIDs, err := qtx.DeleteJournalEntriesForNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	log.With("len", len(journalEntryIDs)).Debug("Deleted debts")

	return tx.Commit()
}

func (p *Persister) CreateUserData(ctx context.Context, namespace string) (
	createJournalEntry func(journalEntry models.ExportedJournalEntry) error,
	createContact func(contact models.ExportedContact) error,
	createDebt func(debt models.ExportedDebt) error,
	createActivity func(activty models.ExportedActivity) error,

	commit func() error,
	rollback func() error,

	err error,
) {
	p.log.With("namespace", namespace).Debug("Creating user data")

	createJournalEntry = func(journalEntry models.ExportedJournalEntry) error { return nil }
	createContact = func(contact models.ExportedContact) error { return nil }
	createDebt = func(debt models.ExportedDebt) error { return nil }
	createActivity = func(activity models.ExportedActivity) error { return nil }

	commit = func() error { return nil }
	rollback = func() error { return nil }

	var tx *sql.Tx
	tx, err = p.db.Begin()
	if err != nil {
		return
	}

	qtx := p.queries.WithTx(tx)

	var (
		contactIDMapLock sync.Mutex
		contactIDMap     = map[int32]int32{}
	)

	createJournalEntry = func(journalEntry models.ExportedJournalEntry) error {
		p.log.With("namespace", namespace).Debug("Creating journal entry", "title", journalEntry.Title, "date", journalEntry.Date, "rating", journalEntry.Rating)

		if _, err := qtx.CreateJournalEntry(ctx, models.CreateJournalEntryParams{
			Title:     journalEntry.Title,
			Body:      journalEntry.Body,
			Rating:    journalEntry.Rating,
			Namespace: namespace,
		}); err != nil {
			return err
		}

		return nil
	}

	createContact = func(contact models.ExportedContact) error {
		p.log.With("namespace", namespace).Debug("Creating contact", "firstName", contact.FirstName, "lastName", contact.LastName, "email", contact.Email)

		c, err := qtx.CreateContact(ctx, models.CreateContactParams{
			FirstName: contact.FirstName,
			LastName:  contact.LastName,
			Nickname:  contact.Nickname,
			Email:     contact.Email,
			Pronouns:  contact.Pronouns,
			Namespace: namespace,
		})
		if err != nil {
			return err
		}

		contactIDMapLock.Lock()
		defer contactIDMapLock.Unlock()

		contactIDMap[contact.ID] = c.ID

		return nil
	}

	createDebt = func(debt models.ExportedDebt) error {
		p.log.With("namespace", namespace).Debug("Creating debt", "amount", debt.Amount, "currency", debt.Currency, "contactID", debt.ContactID)

		contactIDMapLock.Lock()
		defer contactIDMapLock.Unlock()

		if !debt.ContactID.Valid {
			return ErrContactDoesNotExist
		}

		actualContactID, ok := contactIDMap[debt.ContactID.Int32]
		if !ok {
			return ErrContactDoesNotExist
		}

		if _, err := qtx.CreateDebt(ctx, models.CreateDebtParams{
			ID:          actualContactID,
			Amount:      debt.Amount,
			Currency:    debt.Currency,
			Description: debt.Description,
			Namespace:   namespace,
		}); err != nil {
			return err
		}

		return nil
	}

	createActivity = func(activity models.ExportedActivity) error {
		p.log.With("namespace", namespace).Debug("Creating activity", "name", activity.Name, "date", activity.Date, "contactID", activity.ContactID)

		contactIDMapLock.Lock()
		defer contactIDMapLock.Unlock()

		if !activity.ContactID.Valid {
			return ErrContactDoesNotExist
		}

		actualContactID, ok := contactIDMap[activity.ContactID.Int32]
		if !ok {
			return ErrContactDoesNotExist
		}

		if _, err := qtx.CreateActivity(ctx, models.CreateActivityParams{
			ID:          actualContactID,
			Name:        activity.Name,
			Date:        activity.Date,
			Description: activity.Description,
			Namespace:   namespace,
		}); err != nil {
			return err
		}

		return nil
	}

	commit = tx.Commit
	rollback = tx.Rollback

	return
}
