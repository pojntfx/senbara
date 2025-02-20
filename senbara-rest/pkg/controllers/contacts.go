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

	log.Debug("Handling contacts")

	log.Debug("Getting contacts from DB")

	rawContacts, err := c.persister.GetContacts(ctx, namespace)
	if err != nil {
		log.Warn("Could not get contacts from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		return api.GetContacts500TextResponse(errCouldNotFetchFromDB.Error()), nil
	}

	contacts := []api.Contact{}
	for _, rawContact := range rawContacts {
		id := int64(rawContact.ID)

		// TODO: Parse birthday date as date

		contacts = append(contacts, api.Contact{
			Address: &rawContact.Address,
			// Birthday:  rawContact.Birthday.Time,
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
