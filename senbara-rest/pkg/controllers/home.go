package controllers

import (
	"encoding/json"
	"log"
	"net/http"
)

type indexData struct {
	ContactsCount       int64
	JournalEntriesCount int64
}

func (b *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		email, err := b.authorize(r)
		if err != nil {
			log.Println(err)

			http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

			return
		}

		contactsCount, err := b.persister.CountContacts(r.Context(), email)
		if err != nil {
			log.Println(errCouldNotFetchFromDB, err)

			http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

			return
		}

		journalEntriesCount, err := b.persister.CountJournalEntries(r.Context(), email)
		if err != nil {
			log.Println(errCouldNotFetchFromDB, err)

			http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

			return
		}

		if err := json.NewEncoder(w).Encode(indexData{
			ContactsCount:       contactsCount,
			JournalEntriesCount: journalEntriesCount,
		}); err != nil {
			log.Println(errCouldNotWriteResponse, err)

			http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	http.NotFound(w, r)
}
