package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
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
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	c.log.Debug("Exporting user data", "namespace", userData.Email)

	w.Header().Set("Content-Type", "application/jsonl")
	w.Header().Set("Content-Disposition", `attachment; filename="senbara-forms-userdata.jsonl"`)

	encoder := json.NewEncoder(w)

	if err := c.persister.GetUserData(
		r.Context(),

		userData.Email,

		func(journalEntry models.ExportedJournalEntry) error {
			c.log.Debug("Exporting journal entry",
				"journalEntryID", journalEntry.ID,
				"title", journalEntry.Title,
				"date", journalEntry.Date,
				"rating", journalEntry.Rating,
				"namespace", userData.Email,
			)

			journalEntry.ExportedEntityIdentifier.EntityName = EntityNameExportedJournalEntry

			if err := encoder.Encode(journalEntry); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(contact models.ExportedContact) error {
			c.log.Debug("Exporting contact",
				"contactID", contact.ID,
				"firstName", contact.FirstName,
				"lastName", contact.LastName,
				"email", contact.Email,
				"namespace", userData.Email,
			)

			contact.ExportedEntityIdentifier.EntityName = EntityNameExportedContact

			if err := encoder.Encode(contact); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(debt models.ExportedDebt) error {
			c.log.Debug("Exporting debt",
				"debtID", debt.ID,
				"amount", debt.Amount,
				"currency", debt.Currency,
				"contactID", debt.ContactID,
				"namespace", userData.Email,
			)

			debt.ExportedEntityIdentifier.EntityName = EntityNameExportedDebt

			if err := encoder.Encode(debt); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(activity models.ExportedActivity) error {
			c.log.Debug("Exporting activity",
				"activityID", activity.ID,
				"name", activity.Name,
				"date", activity.Date,
				"contactID", activity.ContactID,
				"namespace", userData.Email,
			)

			activity.ExportedEntityIdentifier.EntityName = EntityNameExportedActivity

			if err := encoder.Encode(activity); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
	); err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateUserData(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	c.log.Debug("Starting user data import", "namespace", userData.Email)

	file, _, err := r.FormFile("userData")
	if err != nil {
		log.Println(errCouldNotReadRequest, err)

		http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	createJournalEntry,
		createContact,
		createDebt,
		createActivity,

		commit,
		rollback,

		err := c.persister.CreateUserData(r.Context(), userData.Email)
	if err != nil {
		log.Println(errCouldNotStartTransaction, err)

		http.Error(w, errCouldNotStartTransaction.Error(), http.StatusInternalServerError)

		return
	}
	defer rollback()

	for {
		var b json.RawMessage
		if err := decoder.Decode(&b); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Println(errCouldNotReadRequest, err)

			http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

			return
		}

		var entityIdentifier models.ExportedEntityIdentifier
		if err := json.Unmarshal(b, &entityIdentifier); err != nil {
			log.Println(errCouldNotReadRequest, err)

			http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

			return
		}

		c.log.Debug("Importing user data entity",
			"entityType", entityIdentifier.EntityName,
			"namespace", userData.Email,
		)

		switch entityIdentifier.EntityName {
		case EntityNameExportedJournalEntry:
			var journalEntry models.ExportedJournalEntry
			if err := json.Unmarshal(b, &journalEntry); err != nil {
				log.Println(errCouldNotReadRequest, err)

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			c.log.Debug("Importing journal entry",
				"journalEntryID", journalEntry.ID,
				"title", journalEntry.Title,
				"date", journalEntry.Date,
				"rating", journalEntry.Rating,
				"namespace", userData.Email,
			)

			if err := createJournalEntry(journalEntry); err != nil {
				log.Println(errCouldNotInsertIntoDB, err)

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedContact:
			var contact models.ExportedContact
			if err := json.Unmarshal(b, &contact); err != nil {
				log.Println(errCouldNotReadRequest, err)

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			c.log.Debug("Importing contact",
				"contactID", contact.ID,
				"firstName", contact.FirstName,
				"lastName", contact.LastName,
				"email", contact.Email,
				"namespace", userData.Email,
			)

			if err := createContact(contact); err != nil {
				log.Println(errCouldNotInsertIntoDB, err)

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedDebt:
			var debt models.ExportedDebt
			if err := json.Unmarshal(b, &debt); err != nil {
				log.Println(errCouldNotReadRequest, err)

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			c.log.Debug("Importing debt",
				"debtID", debt.ID,
				"amount", debt.Amount,
				"currency", debt.Currency,
				"contactID", debt.ContactID,
				"namespace", userData.Email,
			)

			if err := createDebt(debt); err != nil {
				log.Println(errCouldNotInsertIntoDB, err)

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		case EntityNameExportedActivity:
			var activity models.ExportedActivity
			if err := json.Unmarshal(b, &activity); err != nil {
				log.Println(errCouldNotReadRequest, err)

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

			c.log.Debug("Importing activity",
				"activityID", activity.ID,
				"name", activity.Name,
				"date", activity.Date,
				"contactID", activity.ContactID,
				"namespace", userData.Email,
			)

			if err := createActivity(activity); err != nil {
				log.Println(errCouldNotInsertIntoDB, err)

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		default:
			c.log.Debug("Skipping user data entity import",
				"err", errUnknownEntityName,
				"entityType", entityIdentifier.EntityName,
				"namespace", userData.Email,
			)

			continue
		}
	}

	c.log.Debug("Completing user data import",
		"namespace", userData.Email,
	)

	if err := commit(); err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (c *Controller) HandleDeleteUserData(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	c.log.Debug("Deleting all user data",
		"namespace", userData.Email,
	)

	if err := c.persister.DeleteUserData(r.Context(), userData.Email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, userData.LogoutURL, http.StatusFound)
}
