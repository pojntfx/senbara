package controllers

import (
	"context"
	"errors"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetIndex(ctx context.Context, request api.GetIndexRequestObject) (api.GetIndexResponseObject, error) {
	namespace := ctx.Value(authn.ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling index")

	log.Debug("Counting contacts and journal entries for index summary")

	contactsAndJournalEntriesCount, err := c.persister.CountContactsAndJournalEntries(ctx, namespace)
	if err != nil {
		log.Warn("Could not count contacts and journal entries for index summary", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetIndex500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	return api.GetIndex200JSONResponse(api.IndexData{
		ContactsCount:       &contactsAndJournalEntriesCount.ContactCount,
		JournalEntriesCount: &contactsAndJournalEntriesCount.JournalEntriesCount,
	}), nil
}
