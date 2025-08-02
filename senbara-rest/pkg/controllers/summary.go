package controllers

import (
	"context"
	"errors"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetSummary(ctx context.Context, request api.GetSummaryRequestObject) (api.GetSummaryResponseObject, error) {
	namespace := ctx.Value(authn.ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling summary")

	log.Debug("Counting contacts and journal entries for summary")

	contactsAndJournalEntriesCount, err := c.persister.CountContactsAndJournalEntries(ctx, namespace)
	if err != nil {
		log.Warn("Could not count contacts and journal entries for summary", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetSummary500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	return api.GetSummary200JSONResponse(api.IndexData{
		ContactsCount:       &contactsAndJournalEntriesCount.ContactCount,
		JournalEntriesCount: &contactsAndJournalEntriesCount.JournalEntriesCount,
	}), nil
}
