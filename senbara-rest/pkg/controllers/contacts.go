package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

func (c *Controller) GetContacts(ctx context.Context, request api.GetContactsRequestObject) (api.GetContactsResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling get contacts")

	rawContacts, err := c.persister.GetContacts(ctx, namespace)
	if err != nil {
		log.Warn("Could not get activity and contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetContacts500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	contacts := []api.Contact{}
	for _, rawContact := range rawContacts {
		id := int64(rawContact.ID)

		var birthday *types.Date
		if rawContact.Birthday.Valid {
			birthday = &types.Date{
				Time: rawContact.Birthday.Time,
			}
		}

		contacts = append(contacts, api.Contact{
			Address:   &rawContact.Address,
			Birthday:  birthday,
			Email:     (*types.Email)(&rawContact.Email),
			FirstName: &rawContact.FirstName,
			Id:        &id,
			LastName:  &rawContact.LastName,
			Nickname:  &rawContact.Nickname,
			Notes:     &rawContact.Notes,
			Pronouns:  &rawContact.Pronouns,
		})
	}

	return api.GetContacts200JSONResponse(contacts), nil
}

func (c *Controller) CreateContact(ctx context.Context, request api.CreateContactRequestObject) (api.CreateContactResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling create contact")

	nickname := ""
	if v := request.Body.Nickname; v != nil {
		nickname = *v
	}

	log.Debug("Creating contact in DB",
		"firstName", request.Body.FirstName,
		"lastName", request.Body.LastName,
		"nickname", nickname,
		"email", request.Body.Email,
		"pronouns", request.Body.Pronouns,
	)

	createdContact, err := c.persister.CreateContact(
		ctx,

		request.Body.FirstName,
		request.Body.LastName,
		nickname,
		string(request.Body.Email),
		request.Body.Pronouns,

		namespace,
	)
	if err != nil {
		log.Warn("Could not create contact in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.CreateContact500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	id := int64(createdContact.ID)

	var birthday *types.Date
	if createdContact.Birthday.Valid {
		birthday = &types.Date{
			Time: createdContact.Birthday.Time,
		}
	}

	return api.CreateContact200JSONResponse{
		Address:   &createdContact.Address,
		Birthday:  birthday,
		Email:     (*types.Email)(&createdContact.Email),
		FirstName: &createdContact.FirstName,
		Id:        &id,
		LastName:  &createdContact.LastName,
		Nickname:  &createdContact.Nickname,
		Notes:     &createdContact.Notes,
		Pronouns:  &createdContact.Pronouns,
	}, nil
}

func (c *Controller) DeleteContact(ctx context.Context, request api.DeleteContactRequestObject) (api.DeleteContactResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling delete contact")

	log.Debug("Deleting contact from DB",
		"id", request.Id,
	)

	id, err := c.persister.DeleteContact(
		ctx,

		int32(request.Id),

		namespace,
	)
	if err != nil {
		log.Warn("Could not delete contact from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		return api.DeleteContact500TextResponse(errCouldNotDeleteFromDB.Error()), nil
	}

	return api.DeleteContact200JSONResponse(id), nil
}

func (c *Controller) GetContact(ctx context.Context, request api.GetContactRequestObject) (api.GetContactResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling get contact")

	log.Debug("Getting contact from DB", "id", request.Id)

	rawContact, err := c.persister.GetContact(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetContact500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	log.Debug("Getting debts for contact from DB", "id", request.Id)

	rawDebts, err := c.persister.GetDebts(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get debts from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetContact500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	log.Debug("Getting activities for contact from DB", "id", request.Id)

	rawActivities, err := c.persister.GetActivities(ctx, int32(request.Id), namespace)
	if err != nil {
		log.Warn("Could not get activities from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetContact500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	activities := []api.Activity{}
	for _, rawActivity := range rawActivities {
		id := int64(rawActivity.ID)

		activities = append(activities, api.Activity{
			Date: &types.Date{
				Time: rawActivity.Date,
			},
			Description: &rawActivity.Description,
			Id:          &id,
			Name:        &rawActivity.Name,
		})
	}

	debts := []api.Debt{}
	for _, rawDebt := range rawDebts {
		id := int64(rawDebt.ID)
		amount := float32(rawDebt.Amount)

		debts = append(debts, api.Debt{
			Amount:      &amount,
			Currency:    &rawDebt.Currency,
			Description: &rawDebt.Description,
			Id:          &id,
		})
	}

	id := int64(rawContact.ID)

	var birthday *types.Date
	if rawContact.Birthday.Valid {
		birthday = &types.Date{
			Time: rawContact.Birthday.Time,
		}
	}

	return api.GetContact200JSONResponse{
		Activities: &activities,
		Debts:      &debts,
		Entry: &api.Contact{
			Address:   &rawContact.Address,
			Birthday:  birthday,
			Email:     (*types.Email)(&rawContact.Email),
			FirstName: &rawContact.FirstName,
			Id:        &id,
			LastName:  &rawContact.LastName,
			Nickname:  &rawContact.Nickname,
			Notes:     &rawContact.Notes,
			Pronouns:  &rawContact.Pronouns,
		},
	}, nil
}

func (c *Controller) UpdateContact(ctx context.Context, request api.UpdateContactRequestObject) (api.UpdateContactResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling update contact")

	nickname := ""
	if v := request.Body.Nickname; v != nil {
		nickname = *v
	}

	address := ""
	if v := request.Body.Address; v != nil {
		address = *v
	}

	notes := ""
	if v := request.Body.Notes; v != nil {
		notes = *v
	}

	var birthday *time.Time
	if request.Body.Birthday != nil {
		birthday = &request.Body.Birthday.Time
	}

	log.Debug("Updating contact in DB",
		"id", request.Id,
		"firstName", request.Body.FirstName,
		"lastName", request.Body.LastName,
		"nickname", nickname,
		"email", request.Body.Email,
		"pronouns", request.Body.Pronouns,
		"birthday", birthday,
		"address", address,
		"notes", notes,
	)

	updatedContact, err := c.persister.UpdateContact(
		ctx,

		int32(request.Id),

		request.Body.FirstName,
		request.Body.LastName,
		nickname,
		string(request.Body.Email),
		request.Body.Pronouns,

		namespace,

		birthday,
		address,
		notes,
	)
	if err != nil {
		log.Warn("Could not update contact in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		return api.UpdateContact500TextResponse(errCouldNotUpdateInDB.Error()), nil
	}

	id := int64(updatedContact.ID)

	var updatedBirthday *types.Date
	if updatedContact.Birthday.Valid {
		updatedBirthday = &types.Date{
			Time: updatedContact.Birthday.Time,
		}
	}

	return api.UpdateContact200JSONResponse{
		Address:   &updatedContact.Address,
		Birthday:  updatedBirthday,
		Email:     (*types.Email)(&updatedContact.Email),
		FirstName: &updatedContact.FirstName,
		Id:        &id,
		LastName:  &updatedContact.LastName,
		Nickname:  &updatedContact.Nickname,
		Notes:     &updatedContact.Notes,
		Pronouns:  &updatedContact.Pronouns,
	}, nil
}
