package controllers

import (
	"context"
	"errors"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetJournalEntries(ctx context.Context, request api.GetJournalEntriesRequestObject) (api.GetJournalEntriesResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling journal")

	log.Debug("Getting journal entries from DB")

	rawJournalEntries, err := c.persister.GetJournalEntries(ctx, namespace)
	if err != nil {
		log.Warn("Could not get journal entries from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetJournalEntries500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	journalEntries := []api.JournalEntry{}
	for _, rawJournalEntry := range rawJournalEntries {
		id := int64(rawJournalEntry.ID)

		journalEntries = append(journalEntries, api.JournalEntry{
			Body:      &rawJournalEntry.Body,
			Date:      &rawJournalEntry.Date,
			Id:        &id,
			Namespace: &rawJournalEntry.Namespace,
			Rating:    &rawJournalEntry.Rating,
			Title:     &rawJournalEntry.Title,
		})
	}

	return api.GetJournalEntries200JSONResponse(journalEntries), nil
}
