package controllers

import (
	"context"
	"errors"

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
	return api.DeleteContact200JSONResponse(0), nil
}

func (c *Controller) GetContact(ctx context.Context, request api.GetContactRequestObject) (api.GetContactResponseObject, error) {
	return api.GetContact200JSONResponse{}, nil
}

func (c *Controller) UpdateContact(ctx context.Context, request api.UpdateContactRequestObject) (api.UpdateContactResponseObject, error) {
	return api.UpdateContact200JSONResponse{}, nil
}
