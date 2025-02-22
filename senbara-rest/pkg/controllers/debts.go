package controllers

import (
	"context"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) CreateDebt(ctx context.Context, request api.CreateDebtRequestObject) (api.CreateDebtResponseObject, error) {
	return api.CreateDebt200JSONResponse{}, nil
}

func (c *Controller) SettleDebt(ctx context.Context, request api.SettleDebtRequestObject) (api.SettleDebtResponseObject, error) {
	return api.SettleDebt200JSONResponse(0), nil
}

func (c *Controller) UpdateDebt(ctx context.Context, request api.UpdateDebtRequestObject) (api.UpdateDebtResponseObject, error) {
	return api.UpdateDebt200JSONResponse{}, nil
}
