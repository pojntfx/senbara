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

func (c *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
	c.log.Debug("Getting index")

	if r.Method == http.MethodGet && r.URL.Path == "/" {
		_, userData, status, err := c.authorize(w, r, false)
		if err != nil {
			log.Println(err)

			http.Error(w, err.Error(), status)

			return
		}

		var contactsCount, journalEntriesCount int64
		if strings.TrimSpace(userData.Email) != "" {
			c.log.Debug("Counting contacts for index summary", "namespace", userData.Email)

			contactsCount, err = c.persister.CountContacts(r.Context(), userData.Email)
			if err != nil {
				log.Println(errCouldNotFetchFromDB, err)

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}

			c.log.Debug("Counting journal entries for index summary", "namespace", userData.Email)

			journalEntriesCount, err = c.persister.CountJournalEntries(r.Context(), userData.Email)
			if err != nil {
				log.Println(errCouldNotFetchFromDB, err)

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}
		}

		if err := c.tpl.ExecuteTemplate(w, "index.html", indexData{
			pageData: pageData{
				userData: userData,

				Page:       userData.Locale.Get("Home"),
				PrivacyURL: c.privacyURL,
				ImprintURL: c.imprintURL,
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

	c.log.Debug("Getting page not found template")

	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	w.WriteHeader(http.StatusNotFound)

	if err := c.tpl.ExecuteTemplate(w, "404.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Page not found"),
		PrivacyURL: c.privacyURL,
		ImprintURL: c.imprintURL,
	},
	); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
