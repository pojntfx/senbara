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
	p.log.Debug("Creating debt", "amount", amount, "currency", currency, "contactID", contactID, "namespace", namespace)

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
	p.log.Debug("Getting debts", "contactID", contactID, "namespace", namespace)

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
	p.log.Debug("Settling debt", "id", id, "namespace", namespace)

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
	p.log.Debug("Getting debt and contact", "id", id, "namespace", namespace)

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
	p.log.Debug("Updating debt", "id", id, "amount", amount, "currency", currency, "namespace", namespace)

	return p.queries.UpdateDebt(ctx, models.UpdateDebtParams{
		ID: id,

		Namespace: namespace,

		Amount:      amount,
		Currency:    currency,
		Description: description,
	})
}
