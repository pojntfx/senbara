package controllers

import (
	"context"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetContacts(ctx context.Context, request api.GetContactsRequestObject) (api.GetContactsResponseObject, error) {
	return api.GetContacts200JSONResponse{}, nil
}

func (c *Controller) CreateContact(ctx context.Context, request api.CreateContactRequestObject) (api.CreateContactResponseObject, error) {
	return api.CreateContact200JSONResponse{}, nil
}

func (c *Controller) DeleteContact(ctx context.Context, request api.DeleteContactRequestObject) (api.DeleteContactResponseObject, error) {
	return api.DeleteContact200JSONResponse(0), nil
}

func (c *Controller) GetContact(ctx context.Context, request api.GetContactRequestObject) (api.GetContactResponseObject, error) {
	return api.GetContact200JSONResponse{}, nil
}

func (c *Controller) UpdateContact(ctx context.Context, request api.UpdateContactRequestObject) (api.UpdateContactResponseObject, error) {
	return api.UpdateContact200JSONResponse{}, nil
}
