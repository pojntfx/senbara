package persisters

import (
	"context"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (p *Persister) GetJournalEntries(ctx context.Context, namespace string) ([]models.JournalEntry, error) {
	p.log.With("namespace", namespace).Debug("Getting journal entries")

	return p.queries.GetJournalEntries(ctx, namespace)
}

func (p *Persister) CreateJournalEntry(ctx context.Context, title, body string, rating int32, namespace string) (models.JournalEntry, error) {
	p.log.With("namespace", namespace).Debug("Creating journal entry", "title", title, "rating", rating)

	return p.queries.CreateJournalEntry(ctx, models.CreateJournalEntryParams{
		Title:     title,
		Body:      body,
		Rating:    rating,
		Namespace: namespace,
	})
}

func (p *Persister) DeleteJournalEntry(ctx context.Context, id int32, namespace string) (int32, error) {
	p.log.With("namespace", namespace).Debug("Deleting journal entry", "id", id)

	return p.queries.DeleteJournalEntry(ctx, models.DeleteJournalEntryParams{
		ID:        id,
		Namespace: namespace,
	})
}

func (p *Persister) GetJournalEntry(ctx context.Context, id int32, namespace string) (models.JournalEntry, error) {
	p.log.With("namespace", namespace).Debug("Getting journal entry", "id", id)

	return p.queries.GetJournalEntry(ctx, models.GetJournalEntryParams{
		ID:        id,
		Namespace: namespace,
	})
}

func (p *Persister) UpdateJournalEntry(ctx context.Context, id int32, title, body string, rating int32, namespace string) (models.JournalEntry, error) {
	p.log.With("namespace", namespace).Debug("Updating journal entry", "id", id, "title", title, "rating", rating)

	return p.queries.UpdateJournalEntry(ctx, models.UpdateJournalEntryParams{
		ID:        id,
		Namespace: namespace,
		Title:     title,
		Body:      body,
		Rating:    rating,
	})
}
