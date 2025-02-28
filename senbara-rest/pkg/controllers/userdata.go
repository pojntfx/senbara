package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
)

const (
	EntityNameExportedJournalEntry = "journalEntry"
	EntityNameExportedContact      = "contact"
	EntityNameExportedDebt         = "debt"
	EntityNameExportedActivity     = "activity"
)

func (c *Controller) DeleteUserData(ctx context.Context, request api.DeleteUserDataRequestObject) (api.DeleteUserDataResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling delete user data")

	log.Debug("Deleting user data from DB")

	if err := c.persister.DeleteUserData(
		ctx,

		namespace,
	); err != nil {
		log.Warn("Could not delete user data from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		return api.DeleteUserData500TextResponse(errCouldNotDeleteFromDB.Error()), nil
	}

	return api.DeleteUserData200Response{}, nil
}

func (c *Controller) ExportUserData(ctx context.Context, request api.ExportUserDataRequestObject) (api.ExportUserDataResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling export user data")

	log.Debug("Getting user data from DB")

	reader, writer := io.Pipe()
	enc := json.NewEncoder(writer)
	go func() {
		defer writer.Close()

		if err := c.persister.GetUserData(
			ctx,

			namespace,

			func(journalEntry models.ExportedJournalEntry) error {
				log.Debug("Exporting journal entry",
					"journalEntryID", journalEntry.ID,
					"title", journalEntry.Title,
					"date", journalEntry.Date,
					"rating", journalEntry.Rating,
				)

				journalEntry.ExportedEntityIdentifier.EntityName = EntityNameExportedJournalEntry

				if err := enc.Encode(journalEntry); err != nil {
					return errors.Join(errCouldNotWriteResponse, err)
				}

				return nil
			},
			func(contact models.ExportedContact) error {
				log.Debug("Exporting contact",
					"contactID", contact.ID,
					"firstName", contact.FirstName,
					"lastName", contact.LastName,
					"email", contact.Email,
				)

				contact.ExportedEntityIdentifier.EntityName = EntityNameExportedContact

				if err := enc.Encode(contact); err != nil {
					return errors.Join(errCouldNotWriteResponse, err)
				}

				return nil
			},
			func(debt models.ExportedDebt) error {
				log.Debug("Exporting debt",
					"debtID", debt.ID,
					"amount", debt.Amount,
					"currency", debt.Currency,
					"contactID", debt.ContactID,
				)

				debt.ExportedEntityIdentifier.EntityName = EntityNameExportedDebt

				if err := enc.Encode(debt); err != nil {
					return errors.Join(errCouldNotWriteResponse, err)
				}

				return nil
			},
			func(activity models.ExportedActivity) error {
				log.Debug("Exporting activity",
					"activityID", activity.ID,
					"name", activity.Name,
					"date", activity.Date,
					"contactID", activity.ContactID,
				)

				activity.ExportedEntityIdentifier.EntityName = EntityNameExportedActivity

				if err := enc.Encode(activity); err != nil {
					return errors.Join(errCouldNotWriteResponse, err)
				}

				return nil
			},
		); err != nil {
			log.Warn("Could not export user data from DB and encode it", "err", errors.Join(errCouldNotFetchFromDB, errCouldNotEncodeResponse, err))

			writer.CloseWithError(errors.Join(errCouldNotFetchFromDB, errCouldNotEncodeResponse))

			return
		}
	}()

	return api.ExportUserData200ApplicationjsonlResponse{
		Body: reader,
		Headers: api.ExportUserData200ResponseHeaders{
			ContentDisposition: `attachment; filename="senbara-forms-userdata.jsonl"`,
		},
	}, nil
}

func (c *Controller) ImportUserData(ctx context.Context, request api.ImportUserDataRequestObject) (api.ImportUserDataResponseObject, error) {
	namespace := ctx.Value(ContextKeyNamespace).(string)

	log := c.log.With("namespace", namespace)

	log.Debug("Handling import user data")

	file, err := request.Body.NextPart()
	if err != nil {
		log.Warn("Could not read user data file from request", "err", errors.Join(errCouldNotReadRequest, err))

		return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
	}
	defer file.Close()

	if file.FormName() != "userData" {
		log.Warn("Could not read user data file from request, invalid file name", "err", errCouldNotReadRequest, "fileName", file.FileName())

		return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
	}

	log.Debug("Importing user data to DB")

	dec := json.NewDecoder(file)

	createJournalEntry,
		createContact,
		createDebt,
		createActivity,

		commit,
		rollback,

		err := c.persister.CreateUserData(ctx, namespace)
	if err != nil {
		log.Warn("Could not start transaction for user data import", "err", errors.Join(errCouldNotStartTransaction, err))

		return api.ImportUserData500TextResponse(errCouldNotStartTransaction.Error()), nil
	}
	defer rollback()

	for {
		var b json.RawMessage
		if err := dec.Decode(&b); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Warn("Could not decode user data", "err", errors.Join(errCouldNotReadRequest, err))

			return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
		}

		var entityIdentifier models.ExportedEntityIdentifier
		if err := json.Unmarshal(b, &entityIdentifier); err != nil {
			log.Warn("Could not unmarshal entity identifier", "err", errors.Join(errCouldNotReadRequest, err))

			return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
		}

		log.Debug("Processing user data entity for user data import", "entityType", entityIdentifier.EntityName)

		switch entityIdentifier.EntityName {
		case EntityNameExportedJournalEntry:
			var journalEntry models.ExportedJournalEntry
			if err := json.Unmarshal(b, &journalEntry); err != nil {
				log.Warn("Could not unmarshal journal entry", "err", errors.Join(errCouldNotReadRequest, err))

				return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
			}

			log.Debug("Importing journal entry",
				"journalEntryID", journalEntry.ID,
				"title", journalEntry.Title,
				"date", journalEntry.Date,
				"rating", journalEntry.Rating,
			)

			if err := createJournalEntry(journalEntry); err != nil {
				log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				return api.ImportUserData500TextResponse(errCouldNotInsertIntoDB.Error()), nil
			}

		case EntityNameExportedContact:
			var contact models.ExportedContact
			if err := json.Unmarshal(b, &contact); err != nil {
				log.Warn("Could not unmarshal contact", "err", errors.Join(errCouldNotReadRequest, err))

				return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
			}

			log.Debug("Importing contact",
				"contactID", contact.ID,
				"firstName", contact.FirstName,
				"lastName", contact.LastName,
				"email", contact.Email,
			)

			if err := createContact(contact); err != nil {
				log.Warn("Could not create contact in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				return api.ImportUserData500TextResponse(errCouldNotInsertIntoDB.Error()), nil
			}

		case EntityNameExportedDebt:
			var debt models.ExportedDebt
			if err := json.Unmarshal(b, &debt); err != nil {
				log.Warn("Could not unmarshal debt", "err", errors.Join(errCouldNotReadRequest, err))

				return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
			}

			log.Debug("Importing debt",
				"debtID", debt.ID,
				"amount", debt.Amount,
				"currency", debt.Currency,
				"contactID", debt.ContactID,
			)

			if err := createDebt(debt); err != nil {
				log.Warn("Could not create debt in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				return api.ImportUserData500TextResponse(errCouldNotInsertIntoDB.Error()), nil
			}

		case EntityNameExportedActivity:
			var activity models.ExportedActivity
			if err := json.Unmarshal(b, &activity); err != nil {
				log.Warn("Could not unmarshal activity", "err", errors.Join(errCouldNotReadRequest, err))

				return api.ImportUserData500TextResponse(errCouldNotReadRequest.Error()), nil
			}

			log.Debug("Importing activity",
				"activityID", activity.ID,
				"name", activity.Name,
				"date", activity.Date,
				"contactID", activity.ContactID,
			)

			if err := createActivity(activity); err != nil {
				log.Warn("Could not create activity in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				return api.ImportUserData500TextResponse(errCouldNotInsertIntoDB.Error()), nil
			}

		default:
			log.Debug("Skipping import of user data entity with unknown entity type",
				"err", errUnknownEntityName,
				"entityType", entityIdentifier.EntityName,
			)

			continue
		}
	}

	log.Debug("Completing user data import")

	if err := commit(); err != nil {
		log.Warn("Could not commit user data import transaction", "err", errors.Join(errCouldNotInsertIntoDB, err))

		return api.ImportUserData500TextResponse(errCouldNotInsertIntoDB.Error()), nil
	}

	return api.ImportUserData200Response{}, nil
}
