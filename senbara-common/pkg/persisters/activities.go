package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (p *Persister) CreateActivity(
	ctx context.Context,

	name string,
	date time.Time,
	description string,

	contactID int32,
	namespace string,
) (int32, error) {
	return p.queries.CreateActivity(ctx, models.CreateActivityParams{
		ID:          contactID,
		Namespace:   namespace,
		Name:        name,
		Date:        date,
		Description: description,
	})
}

func (p *Persister) GetActivities(
	ctx context.Context,

	contactID int32,
	namespace string,
) ([]models.GetActivitiesRow, error) {
	return p.queries.GetActivities(ctx, models.GetActivitiesParams{
		ID:        contactID,
		Namespace: namespace,
	})
}

func (p *Persister) DeleteActivity(
	ctx context.Context,

	id int32,

	namespace string,
) error {
	return p.queries.DeleteActivity(ctx, models.DeleteActivityParams{
		ID: id,

		Namespace: namespace,
	})
}

func (p *Persister) GetActivityAndContact(
	ctx context.Context,

	id int32,

	namespace string,
) (models.GetActivityAndContactRow, error) {
	return p.queries.GetActivityAndContact(ctx, models.GetActivityAndContactParams{
		ID: id,

		Namespace: namespace,
	})
}

func (p *Persister) UpdateActivity(
	ctx context.Context,

	id int32,

	namespace string,

	name string,
	date time.Time,
	description string,
) error {
	return p.queries.UpdateActivity(ctx, models.UpdateActivityParams{
		ID: id,

		Namespace: namespace,

		Name:        name,
		Date:        date,
		Description: description,
	})
}
