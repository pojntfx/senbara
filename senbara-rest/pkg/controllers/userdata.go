package controllers

import (
	"context"

	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) DeleteUserData(ctx context.Context, request api.DeleteUserDataRequestObject) (api.DeleteUserDataResponseObject, error) {
	return api.DeleteUserData200Response{}, nil
}

func (c *Controller) ExportUserData(ctx context.Context, request api.ExportUserDataRequestObject) (api.ExportUserDataResponseObject, error) {
	return api.ExportUserData200ApplicationjsonlResponse{
		Headers: api.ExportUserData200ResponseHeaders{
			ContentDisposition: `attachment; filename="senbara-forms-userdata.jsonl"`,
		},
	}, nil
}

func (c *Controller) ImportUserData(ctx context.Context, request api.ImportUserDataRequestObject) (api.ImportUserDataResponseObject, error) {
	return api.ImportUserData200Response{}, nil
}
