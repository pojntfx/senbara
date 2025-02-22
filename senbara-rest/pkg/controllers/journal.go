package controllers

import (
	"context"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetJournalEntries(ctx context.Context, request api.GetJournalEntriesRequestObject) (api.GetJournalEntriesResponseObject, error) {
	return api.GetJournalEntries200JSONResponse{}, nil
}

func (c *Controller) CreateJournalEntry(ctx context.Context, request api.CreateJournalEntryRequestObject) (api.CreateJournalEntryResponseObject, error) {
	return api.CreateJournalEntry200JSONResponse{}, nil
}

func (c *Controller) DeleteJournalEntry(ctx context.Context, request api.DeleteJournalEntryRequestObject) (api.DeleteJournalEntryResponseObject, error) {
	return api.DeleteJournalEntry200JSONResponse(0), nil
}

func (c *Controller) GetJournalEntry(ctx context.Context, request api.GetJournalEntryRequestObject) (api.GetJournalEntryResponseObject, error) {
	return api.GetJournalEntry200JSONResponse{}, nil
}

func (c *Controller) UpdateJournalEntry(ctx context.Context, request api.UpdateJournalEntryRequestObject) (api.UpdateJournalEntryResponseObject, error) {
	return api.UpdateJournalEntry200JSONResponse{}, nil
}
