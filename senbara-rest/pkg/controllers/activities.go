package controllers

import (
	"context"
	"errors"

	"github.com/oapi-codegen/runtime/types"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) CreateActivity(ctx context.Context, request api.CreateActivityRequestObject) (api.CreateActivityResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling create activity")

	description := ""
	if v := request.Body.Description; v != nil {
		description = *v
	}

	log.Debug("Creating activity in DB",
		"contactID", request.Body.ContactId,
		"name", request.Body.Name,
		"date", request.Body.Date,
		"description", request.Body.Description,
	)

	createdActivity, err := c.persister.CreateActivity(
		ctx,

		request.Body.Name,
		request.Body.Date.Time,
		description,

		int32(request.Body.ContactId),
		namespace,
	)
	if err != nil {
		log.Warn("Could not create activity in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.CreateActivity500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	id := int64(createdActivity.ID)

	return api.CreateActivity200JSONResponse{
		Date: &types.Date{
			Time: createdActivity.Date,
		},
		Description: &createdActivity.Description,
		Id:          &id,
		Name:        &createdActivity.Name,
	}, nil
}

func (c *Controller) DeleteActivity(ctx context.Context, request api.DeleteActivityRequestObject) (api.DeleteActivityResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling delete activity")

	log.Debug("Deleting activity from DB",
		"id", request.Id,
	)

	id, err := c.persister.DeleteActivity(
		ctx,

		int32(request.Id),

		namespace,
	)
	if err != nil {
		log.Warn("Could not delete activity from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		return api.DeleteActivity500TextResponse(errCouldNotDeleteFromDB.Error()), nil
	}

	return api.DeleteActivity200JSONResponse(id), nil
}

func (c *Controller) GetActivity(ctx context.Context, request api.GetActivityRequestObject) (api.GetActivityResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling get activity")

	log.Debug("Getting activity from DB",
		"id", request.Id,
	)

	activityAndContact, err := c.persister.GetActivityAndContact(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get activity and contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetActivity500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	activityID := int64(activityAndContact.ActivityID)
	contactID := int64(activityAndContact.ContactID)

	return api.GetActivity200JSONResponse{
		ActivityId: &activityID,
		ContactId:  &contactID,
		Date: &types.Date{
			Time: activityAndContact.Date,
		},
		Description: &activityAndContact.Description,
		FirstName:   &activityAndContact.FirstName,
		LastName:    &activityAndContact.LastName,
		Name:        &activityAndContact.Name,
	}, nil
}

func (c *Controller) UpdateActivity(ctx context.Context, request api.UpdateActivityRequestObject) (api.UpdateActivityResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling update activity")

	description := ""
	if v := request.Body.Description; v != nil {
		description = *v
	}

	log.Debug("Updating activity in DB",
		"id", request.Id,
		"name", request.Body.Name,
		"date", request.Body.Date.Time,
		"description", description,
	)

	updatedActivity, err := c.persister.UpdateActivity(
		ctx,

		int32(request.Id),

		namespace,

		request.Body.Name,
		request.Body.Date.Time,
		description,
	)
	if err != nil {
		log.Warn("Could not update activity in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		return api.UpdateActivity500TextResponse(errCouldNotUpdateInDB.Error()), nil
	}

	id := int64(updatedActivity.ID)

	return api.UpdateActivity200JSONResponse{
		Date: &types.Date{
			Time: updatedActivity.Date,
		},
		Description: &updatedActivity.Description,
		Id:          &id,
		Name:        &updatedActivity.Name,
	}, nil
}
