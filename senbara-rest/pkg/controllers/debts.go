package controllers

import (
	"context"
	"errors"
	"math"

	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) CreateDebt(ctx context.Context, request api.CreateDebtRequestObject) (api.CreateDebtResponseObject, error) {
	namespace := ctx.Value(authn.ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling create debt")

	amount := math.Abs(float64(request.Body.Amount))
	if request.Body.YouOwe {
		amount = -amount
	}

	description := ""
	if v := request.Body.Description; v != nil {
		description = *v
	}

	log.Debug("Creating debt in DB",
		"contactID", request.Body.ContactId,
		"amount", amount,
		"currency", request.Body.Currency,
		"description", description,
	)

	createdDebt, err := c.persister.CreateDebt(
		ctx,

		amount,
		request.Body.Currency,
		description,
		int32(request.Body.ContactId),

		namespace,
	)
	if err != nil {
		log.Warn("Could not create deb in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.CreateDebt500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	id := int64(createdDebt.ID)
	updatedAmount := float32(createdDebt.Amount)

	return api.CreateDebt200JSONResponse{
		Amount:      &updatedAmount,
		Currency:    &createdDebt.Currency,
		Description: &createdDebt.Description,
		Id:          &id,
	}, nil
}

func (c *Controller) SettleDebt(ctx context.Context, request api.SettleDebtRequestObject) (api.SettleDebtResponseObject, error) {
	namespace := ctx.Value(authn.ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling settle debt")

	log.Debug("Settling debt in DB",
		"id", request.Id,
	)

	id, err := c.persister.SettleDebt(
		ctx,

		int32(request.Id),

		namespace,
	)
	if err != nil {
		log.Warn("Could not settle debt in DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		return api.SettleDebt500TextResponse(errCouldNotDeleteFromDB.Error()), nil
	}

	return api.SettleDebt200JSONResponse(id), nil
}

func (c *Controller) UpdateDebt(ctx context.Context, request api.UpdateDebtRequestObject) (api.UpdateDebtResponseObject, error) {
	namespace := ctx.Value(authn.ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling update debt")

	amount := math.Abs(float64(request.Body.Amount))
	if request.Body.YouOwe {
		amount = -amount
	}

	description := ""
	if v := request.Body.Description; v != nil {
		description = *v
	}

	log.Debug("Updating debt in DB",
		"id", request.Id,
		"amount", amount,
		"currency", request.Body.Currency,
		"description", description,
	)

	updatedDebt, err := c.persister.UpdateDebt(
		ctx,

		int32(request.Id),

		namespace,

		amount,
		request.Body.Currency,
		description,
	)
	if err != nil {
		log.Warn("Could not update debt in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		return api.UpdateDebt500TextResponse(errCouldNotUpdateInDB.Error()), nil
	}

	id := int64(updatedDebt.ID)
	updatedAmount := float32(updatedDebt.Amount)

	return api.UpdateDebt200JSONResponse{
		Amount:      &updatedAmount,
		Currency:    &updatedDebt.Currency,
		Description: &updatedDebt.Description,
		Id:          &id,
	}, nil
}
