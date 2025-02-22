package controllers

import (
	"context"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) CreateActivity(ctx context.Context, request api.CreateActivityRequestObject) (api.CreateActivityResponseObject, error) {
	return api.CreateActivity200JSONResponse{}, nil
}

func (c *Controller) DeleteActivity(ctx context.Context, request api.DeleteActivityRequestObject) (api.DeleteActivityResponseObject, error) {
	return api.DeleteActivity200JSONResponse(0), nil
}

func (c *Controller) GetActivity(ctx context.Context, request api.GetActivityRequestObject) (api.GetActivityResponseObject, error) {
	return api.GetActivity200JSONResponse{}, nil
}

func (c *Controller) UpdateActivity(ctx context.Context, request api.UpdateActivityRequestObject) (api.UpdateActivityResponseObject, error) {
	return api.UpdateActivity200JSONResponse{}, nil
}
