package persisters

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/pojntfx/senbara/senbara-forms/pkg/models"
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
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := p.queries.WithTx(tx)

	if err := qtx.DeleteActivitiesForNamespace(ctx, namespace); err != nil {
		return err
	}

	if err := qtx.DeleteDebtsForNamespace(ctx, namespace); err != nil {
		return err
	}

	if err := qtx.DeleteContactsForNamespace(ctx, namespace); err != nil {
		return err
	}

	if err := qtx.DeleteJournalEntriesForNamespace(ctx, namespace); err != nil {
		return err
	}

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
		journalEntryIDMapLock sync.Mutex
		journalEntryIDMap     = map[int32]int32{}
	)

	createJournalEntry = func(journalEntry models.ExportedJournalEntry) error {
		id, err := qtx.CreateJournalEntry(ctx, models.CreateJournalEntryParams{
			Title:  journalEntry.Title,
			Body:   journalEntry.Body,
			Rating: journalEntry.Rating,

			Namespace: namespace,
		})
		if err != nil {
			return err
		}

		journalEntryIDMapLock.Lock()
		defer journalEntryIDMapLock.Unlock()

		journalEntryIDMap[journalEntry.ID] = id

		return nil
	}

	// TODO: Import debts and activities, use `journalEntryIDMap` to resolve external
	// to internal/actual IDs

	createContact = func(contact models.ExportedContact) error {
		if _, err := qtx.CreateContact(ctx, models.CreateContactParams{
			FirstName: contact.FirstName,
			LastName:  contact.LastName,
			Nickname:  contact.Nickname,
			Email:     contact.Email,
			Pronouns:  contact.Pronouns,

			Namespace: namespace,
		}); err != nil {
			return err
		}

		return nil
	}

	createDebt = func(debt models.ExportedDebt) error {
		journalEntryIDMapLock.Lock()
		defer journalEntryIDMapLock.Unlock()

		if !debt.ContactID.Valid {
			return ErrContactDoesNotExist
		}

		actualContactID, ok := journalEntryIDMap[debt.ContactID.Int32]
		if !ok {
			return ErrContactDoesNotExist
		}

		if _, err := qtx.CreateDebt(ctx, models.CreateDebtParams{
			ID:          actualContactID,
			Amount:      debt.Amount,
			Currency:    debt.Currency,
			Description: debt.Description,

			Namespace: namespace,
		}); err != nil {
			return err
		}

		return nil
	}

	commit = tx.Commit
	rollback = tx.Rollback

	return
}
