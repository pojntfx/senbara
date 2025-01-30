package persisters

import (
	"context"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (p *Persister) CreateDebt(
	ctx context.Context,

	amount float64,
	currency,
	description string,

	contactID int32,
	namespace string,
) (int32, error) {
	return p.queries.CreateDebt(ctx, models.CreateDebtParams{
		ID:          contactID,
		Namespace:   namespace,
		Amount:      amount,
		Currency:    currency,
		Description: description,
	})
}

func (p *Persister) GetDebts(
	ctx context.Context,

	contactID int32,
	namespace string,
) ([]models.GetDebtsRow, error) {
	return p.queries.GetDebts(ctx, models.GetDebtsParams{
		ID:        contactID,
		Namespace: namespace,
	})
}

func (p *Persister) SettleDebt(
	ctx context.Context,

	id int32,

	namespace string,
) error {
	return p.queries.SettleDebt(ctx, models.SettleDebtParams{
		ID: id,

		Namespace: namespace,
	})
}

func (p *Persister) GetDebtAndContact(
	ctx context.Context,

	id int32,

	namespace string,
) (models.GetDebtAndContactRow, error) {
	return p.queries.GetDebtAndContact(ctx, models.GetDebtAndContactParams{
		ID: id,

		Namespace: namespace,
	})
}

func (p *Persister) UpdateDebt(
	ctx context.Context,

	id int32,

	namespace string,

	amount float64,
	currency,
	description string,
) error {
	return p.queries.UpdateDebt(ctx, models.UpdateDebtParams{
		ID: id,

		Namespace: namespace,

		Amount:      amount,
		Currency:    currency,
		Description: description,
	})
}
