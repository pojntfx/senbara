package controllers

import (
	"log"
	"net/http"
	"strings"
)

type indexData struct {
	pageData
	ContactsCount       int64
	JournalEntriesCount int64
}

func (b *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		_, userData, status, err := b.authorize(w, r, false)
		if err != nil {
			log.Println(err)

			http.Error(w, err.Error(), status)

			return
		}

		var contactsCount, journalEntriesCount int64
		if strings.TrimSpace(userData.Email) != "" {
			contactsCount, err = b.persister.CountContacts(r.Context(), userData.Email)
			if err != nil {
				log.Println(errCouldNotFetchFromDB, err)

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}

			journalEntriesCount, err = b.persister.CountJournalEntries(r.Context(), userData.Email)
			if err != nil {
				log.Println(errCouldNotFetchFromDB, err)

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}
		}

		if err := b.tpl.ExecuteTemplate(w, "index.html", indexData{
			pageData: pageData{
				userData: userData,

				Page:       userData.Locale.Get("Home"),
				PrivacyURL: b.privacyURL,
				ImprintURL: b.imprintURL,
			},
			ContactsCount:       contactsCount,
			JournalEntriesCount: journalEntriesCount,
		}); err != nil {
			log.Println(errCouldNotRenderTemplate, err)

			http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := b.tpl.ExecuteTemplate(w, "404.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Page not found"),
		PrivacyURL: b.privacyURL,
		ImprintURL: b.imprintURL,
	},
	); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
