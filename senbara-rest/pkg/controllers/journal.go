package controllers

import (
	"context"
	"errors"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetJournalEntries(ctx context.Context, request api.GetJournalEntriesRequestObject) (api.GetJournalEntriesResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling get journal entries")

	rawEntries, err := c.persister.GetJournalEntries(ctx, namespace)
	if err != nil {
		log.Warn("Could not get journal entries from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetJournalEntries500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	entries := []api.JournalEntry{}
	for _, rawEntry := range rawEntries {
		id := int64(rawEntry.ID)

		entries = append(entries, api.JournalEntry{
			Body:   &rawEntry.Body,
			Date:   &rawEntry.Date,
			Id:     &id,
			Rating: &rawEntry.Rating,
			Title:  &rawEntry.Title,
		})
	}

	return api.GetJournalEntries200JSONResponse(entries), nil
}

func (c *Controller) CreateJournalEntry(ctx context.Context, request api.CreateJournalEntryRequestObject) (api.CreateJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling create journal entry")

	log.Debug("Creating journal entry in DB",
		"title", request.Body.Title,
		"rating", request.Body.Body,
	)

	createdJournalEntry, err := c.persister.CreateJournalEntry(
		ctx,

		request.Body.Title,
		request.Body.Body,
		request.Body.Rating,

		namespace,
	)
	if err != nil {
		log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.CreateJournalEntry500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	id := int64(createdJournalEntry.ID)

	return api.CreateJournalEntry200JSONResponse{
		Body:   &createdJournalEntry.Body,
		Date:   &createdJournalEntry.Date,
		Id:     &id,
		Rating: &createdJournalEntry.Rating,
		Title:  &createdJournalEntry.Title,
	}, nil
}

func (c *Controller) DeleteJournalEntry(ctx context.Context, request api.DeleteJournalEntryRequestObject) (api.DeleteJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling delete journal entry")

	log.Debug("Deleting journal entry from DB",
		"id", request.Id,
	)

	id, err := c.persister.DeleteJournalEntry(
		ctx,

		int32(request.Id),

		namespace,
	)
	if err != nil {
		log.Warn("Could not delete journal entry from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		return api.DeleteJournalEntry500TextResponse(errCouldNotDeleteFromDB.Error()), nil
	}

	return api.DeleteJournalEntry200JSONResponse(id), nil
}

func (c *Controller) GetJournalEntry(ctx context.Context, request api.GetJournalEntryRequestObject) (api.GetJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling get journal entry")

	log.Debug("Getting journal entry from DB",
		"id", request.Id,
	)

	rawJournalEntry, err := c.persister.GetJournalEntry(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetJournalEntry500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	id := int64(rawJournalEntry.ID)

	return api.GetJournalEntry200JSONResponse{
		Body:   &rawJournalEntry.Body,
		Date:   &rawJournalEntry.Date,
		Id:     &id,
		Rating: &rawJournalEntry.Rating,
		Title:  &rawJournalEntry.Title,
	}, nil
}

func (c *Controller) UpdateJournalEntry(ctx context.Context, request api.UpdateJournalEntryRequestObject) (api.UpdateJournalEntryResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling update journal entry")

	log.Debug("Updating journal entry in DB",
		"id", request.Id,
		"title", request.Body.Title,
		"rating", request.Body.Rating,
	)

	updatedJournalEntry, err := c.persister.UpdateJournalEntry(
		ctx,

		int32(request.Id),
		request.Body.Title,
		request.Body.Body,
		request.Body.Rating,

		namespace,
	)
	if err != nil {
		log.Warn("Could not update journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.UpdateJournalEntry500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	id := int64(updatedJournalEntry.ID)

	return api.UpdateJournalEntry200JSONResponse{
		Body:   &updatedJournalEntry.Body,
		Date:   &updatedJournalEntry.Date,
		Id:     &id,
		Rating: &updatedJournalEntry.Rating,
		Title:  &updatedJournalEntry.Title,
	}, nil
}
