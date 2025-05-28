package controllers

import (
	"errors"
	"net/http"
	"strings"
)

type indexData struct {
	pageData
	ContactsCount       int64
	JournalEntriesCount int64
}

func (c *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
	c.log.Debug("Handling index")

	if r.Method == http.MethodGet && r.URL.Path == "/" {
		_, userData, status, err := c.authorize(w, r, false)
		if err != nil {
			c.log.Warn("Could not authorize user for index page", "err", err)

			http.Error(w, err.Error(), status)

			return
		}

		var contactsCount, journalEntriesCount int64
		if strings.TrimSpace(userData.Email) != "" {
			log := c.log.With("namespace", userData.Email)

			log.Debug("Counting contacts for index summary")

			var err error
			contactsCount, err = c.persister.CountContacts(r.Context(), userData.Email)
			if err != nil {
				log.Warn("Could not count contacts for index summary", "err", errors.Join(errCouldNotFetchFromDB, err))

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}

			log.Debug("Counting journal entries for index summary")

			journalEntriesCount, err = c.persister.CountJournalEntries(r.Context(), userData.Email)
			if err != nil {
				log.Warn("Could not count journal entries for index summary", "err", errors.Join(errCouldNotFetchFromDB, err))

				http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

				return
			}
		}

		if err := c.tpl.ExecuteTemplate(w, "index.html", indexData{
			pageData: pageData{
				userData: userData,

				Page:       userData.Locale.Get("Home"),
				PrivacyURL: c.privacyURL,
				TosURL:     c.tosURL,
				ImprintURL: c.imprintURL,
			},
			ContactsCount:       contactsCount,
			JournalEntriesCount: journalEntriesCount,
		}); err != nil {
			c.log.Warn("Could not render index template", "err", errors.Join(errCouldNotRenderTemplate, err))

			http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	c.log.Debug("Handling page not found")

	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for page not found page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	w.WriteHeader(http.StatusNotFound)

	if err := c.tpl.ExecuteTemplate(w, "404.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Page not found"),
		PrivacyURL: c.privacyURL,
		TosURL:     c.tosURL,
		ImprintURL: c.imprintURL,
	}); err != nil {
		log.Warn("Could not render page not found template", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
