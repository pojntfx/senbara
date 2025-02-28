package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

const (
	EntityNameExportedJournalEntry = "journalEntry"
	EntityNameExportedContact      = "contact"
	EntityNameExportedDebt         = "debt"
	EntityNameExportedActivity     = "activity"
)

func (c *Controller) HandleUserData(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for user data export", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling export user data")

	w.Header().Set("Content-Type", "application/jsonl")
	w.Header().Set("Content-Disposition", `attachment; filename="senbara-forms-userdata.jsonl"`)

	log.Debug("Getting user data from DB")

	enc := json.NewEncoder(w)

	if err := c.persister.GetUserData(
		r.Context(),

		userData.Email,

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

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateUserData(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for user data import", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling import user data")

	file, _, err := r.FormFile("userData")
	if err != nil {
		log.Warn("Could not read user data file from request", "err", errors.Join(errCouldNotReadRequest, err))

		http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

		return
	}
	defer file.Close()

	log.Debug("Importing user data to DB")

	dec := json.NewDecoder(file)

	createJournalEntry,
		createContact,
		createDebt,
		createActivity,

		commit,
		rollback,

		err := c.persister.CreateUserData(r.Context(), userData.Email)
	if err != nil {
		log.Warn("Could not start transaction for user data import", "err", errors.Join(errCouldNotStartTransaction, err))

		http.Error(w, errCouldNotStartTransaction.Error(), http.StatusInternalServerError)

		return
	}
	defer rollback()

	for {
		var b json.RawMessage
		if err := dec.Decode(&b); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Warn("Could not decode user data", "err", errors.Join(errCouldNotReadRequest, err))

			http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

			return
		}

		var entityIdentifier models.ExportedEntityIdentifier
		if err := json.Unmarshal(b, &entityIdentifier); err != nil {
			log.Warn("Could not unmarshal entity identifier", "err", errors.Join(errCouldNotReadRequest, err))

			http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

			return
		}

		log.Debug("Processing user data entity for user data import", "entityType", entityIdentifier.EntityName)

		switch entityIdentifier.EntityName {
		case EntityNameExportedJournalEntry:
			var journalEntry models.ExportedJournalEntry
			if err := json.Unmarshal(b, &journalEntry); err != nil {
				log.Warn("Could not unmarshal journal entry", "err", errors.Join(errCouldNotReadRequest, err))

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			log.Debug("Importing journal entry",
				"journalEntryID", journalEntry.ID,
				"title", journalEntry.Title,
				"date", journalEntry.Date,
				"rating", journalEntry.Rating,
			)

			if err := createJournalEntry(journalEntry); err != nil {
				log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedContact:
			var contact models.ExportedContact
			if err := json.Unmarshal(b, &contact); err != nil {
				log.Warn("Could not unmarshal contact", "err", errors.Join(errCouldNotReadRequest, err))

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			log.Debug("Importing contact",
				"contactID", contact.ID,
				"firstName", contact.FirstName,
				"lastName", contact.LastName,
				"email", contact.Email,
			)

			if err := createContact(contact); err != nil {
				log.Warn("Could not create contact in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedDebt:
			var debt models.ExportedDebt
			if err := json.Unmarshal(b, &debt); err != nil {
				log.Warn("Could not unmarshal debt", "err", errors.Join(errCouldNotReadRequest, err))

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			log.Debug("Importing debt",
				"debtID", debt.ID,
				"amount", debt.Amount,
				"currency", debt.Currency,
				"contactID", debt.ContactID,
			)

			if err := createDebt(debt); err != nil {
				log.Warn("Could not create debt in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedActivity:
			var activity models.ExportedActivity
			if err := json.Unmarshal(b, &activity); err != nil {
				log.Warn("Could not unmarshal activity", "err", errors.Join(errCouldNotReadRequest, err))

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			log.Debug("Importing activity",
				"activityID", activity.ID,
				"name", activity.Name,
				"date", activity.Date,
				"contactID", activity.ContactID,
			)

			if err := createActivity(activity); err != nil {
				log.Warn("Could not create activity in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
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

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (c *Controller) HandleDeleteUserData(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for user data deletion", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling delete user data")

	log.Debug("Deleting user data from DB")

	if err := c.persister.DeleteUserData(r.Context(), userData.Email); err != nil {
		log.Warn("Could not delete user data from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, userData.LogoutURL, http.StatusFound)
}
