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

func (c *Controller) CreateJournalEntry(ctx context.Context, request api.CreateJournalEntryRequestObject) (api.CreateJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling create journal")

	log.Debug("Creating journal entry in DB",
		"title", request.Body.Title,
		"rating", request.Body.Rating,
	)

	id, err := c.persister.CreateJournalEntry(ctx, request.Body.Title, request.Body.Body, request.Body.Rating, namespace)
	if err != nil {
		log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.CreateJournalEntry500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	return api.CreateJournalEntry200JSONResponse(id), nil
}

func (c *Controller) GetJournalEntry(ctx context.Context, request api.GetJournalEntryRequestObject) (api.GetJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling view journal")

	log.Debug("Getting journal entry from DB",
		"id",
		request.Id,
	)

	journalEntry, err := c.persister.GetJournalEntry(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get journal entry from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetJournalEntry500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	id := int64(journalEntry.ID)

	return api.GetJournalEntry200JSONResponse(api.JournalEntry{
		Body:      &journalEntry.Body,
		Date:      &journalEntry.Date,
		Id:        &id,
		Namespace: &journalEntry.Namespace,
		Rating:    &journalEntry.Rating,
		Title:     &journalEntry.Title,
	}), nil
}

func (c *Controller) DeleteJournalEntry(ctx context.Context, request api.DeleteJournalEntryRequestObject) (api.DeleteJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling delete journal")

	log.Debug("Deleting journal entry from DB",
		"id", request.Id,
	)

	if err := c.persister.DeleteJournalEntry(ctx, int32(request.Id), namespace); err != nil {
		log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.DeleteJournalEntry500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	return api.DeleteJournalEntry200JSONResponse(request.Id), nil
}
