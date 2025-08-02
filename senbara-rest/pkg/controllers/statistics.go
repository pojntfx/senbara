package controllers

import (
	"context"
	"errors"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetStatistics(ctx context.Context, request api.GetStatisticsRequestObject) (api.GetStatisticsResponseObject, error) {
	c.log.Debug("Handling statistics")

	c.log.Debug("Counting all contacts and journal entries for statistics")

	contactsAndJournalEntriesCount, err := c.persister.CountAllContactsAndJournalEntries(ctx)
	if err != nil {
		c.log.Warn("Could not count all contacts and journal entries for statistics", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetStatistics500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	return api.GetStatistics200JSONResponse(api.IndexData{
		ContactsCount:       &contactsAndJournalEntriesCount.ContactCount,
		JournalEntriesCount: &contactsAndJournalEntriesCount.JournalEntriesCount,
	}), nil
}
