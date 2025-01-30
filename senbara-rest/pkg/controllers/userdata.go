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

func (b *Controller) HandleUserData(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	w.Header().Set("Content-Type", "application/jsonl")
	w.Header().Set("Content-Disposition", `attachment; filename="senbara-forms-userdata.jsonl"`)

	encoder := json.NewEncoder(w)

	if err := b.persister.GetUserData(
		r.Context(),

		email,

		func(journalEntry models.ExportedJournalEntry) error {
			journalEntry.ExportedEntityIdentifier.EntityName = EntityNameExportedJournalEntry

			if err := encoder.Encode(journalEntry); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(contact models.ExportedContact) error {
			contact.ExportedEntityIdentifier.EntityName = EntityNameExportedContact

			if err := encoder.Encode(contact); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(debt models.ExportedDebt) error {
			debt.ExportedEntityIdentifier.EntityName = EntityNameExportedDebt

			if err := encoder.Encode(debt); err != nil {
				return errors.Join(errCouldNotWriteResponse, err)
			}

			return nil
		},
		func(activity models.ExportedActivity) error {
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

func (b *Controller) HandleCreateUserData(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

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

		err := b.persister.CreateUserData(r.Context(), email)
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

		switch entityIdentifier.EntityName {
		case EntityNameExportedJournalEntry:
			var journalEntry models.ExportedJournalEntry
			if err := json.Unmarshal(b, &journalEntry); err != nil {
				log.Println(errCouldNotReadRequest, err)

				http.Error(w, errCouldNotReadRequest.Error(), http.StatusInternalServerError)

				return
			}

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

			if err := createActivity(activity); err != nil {
				log.Println(errCouldNotInsertIntoDB, err)

				http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

				return
			}

		default:
			log.Println("Skipping import error:", errUnknownEntityName, err)

			continue
		}
	}

	if err := commit(); err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleDeleteUserData(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	if err := b.persister.DeleteUserData(r.Context(), email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}
}
