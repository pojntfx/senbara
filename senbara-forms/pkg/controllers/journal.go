package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

type journalData struct {
	pageData
	Entries []models.JournalEntry
}

type journalEntryData struct {
	pageData
	Entry models.JournalEntry
}

func (c *Controller) HandleJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for journal page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling journal page")

	journalEntries, err := c.persister.GetJournalEntries(r.Context(), userData.Email)
	if err != nil {
		log.Warn("Could not get journal entries from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "journal.html", journalData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Journal"),
			PrivacyURL: c.privacyURL,
			TosURL:     c.tosURL,
			ImprintURL: c.imprintURL,
		},
		Entries: journalEntries,
	}); err != nil {
		log.Warn("Could not render template for journal page", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleAddJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for add journal page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling add journal page")

	if err := c.tpl.ExecuteTemplate(w, "journal_add.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Add a journal entry"),
		PrivacyURL: c.privacyURL,
		TosURL:     c.tosURL,
		ImprintURL: c.imprintURL,
	}); err != nil {
		log.Warn("Could not render template for adding journal entry", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for create journal", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling create journal")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not create journal entry", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		log.Warn("Could not create journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	body := r.FormValue("body")
	if strings.TrimSpace(body) == "" {
		log.Warn("Could not create journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rrating := r.FormValue("rating")
	if strings.TrimSpace(rrating) == "" {
		log.Warn("Could not create journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rating, err := strconv.Atoi(rrating)
	if err != nil {
		log.Warn("Could not create journal entry", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Creating journal entry in DB",
		"title", title,
		"rating", rating,
	)

	createdJournalEntry, err := c.persister.CreateJournalEntry(r.Context(), title, body, int32(rating), userData.Email)
	if err != nil {
		log.Warn("Could not create journal entry in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/journal/view?id=%v", createdJournalEntry.ID), http.StatusFound)
}

func (c *Controller) HandleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for delete journal", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling delete journal")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not delete journal entry", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not delete journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not delete journal entry", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Deleting journal entry from DB",
		"id", id,
	)

	if _, err := c.persister.DeleteJournalEntry(r.Context(), int32(id), userData.Email); err != nil {
		log.Warn("Could not delete journal entry from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/journal", http.StatusFound)
}

func (c *Controller) HandleEditJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for edit journal page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling edit journal page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare edit journal page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare edit journal page", "err", errors.Join(errInvalidQueryParam, err))

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting journal entry for edit",
		"id", id,
	)

	journalEntry, err := c.persister.GetJournalEntry(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get journal entry for edit from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "journal_edit.html", journalEntryData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit journal entry"),
			PrivacyURL: c.privacyURL,
			TosURL:     c.tosURL,
			ImprintURL: c.imprintURL,
		},
		Entry: journalEntry,
	}); err != nil {
		log.Warn("Could not render template for editing journal entry", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleUpdateJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for update journal", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling update journal")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not update journal entry", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not update journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not update journal entry", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		log.Warn("Could not update journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	body := r.FormValue("body")
	if strings.TrimSpace(body) == "" {
		log.Warn("Could not update journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rrating := r.FormValue("rating")
	if strings.TrimSpace(rrating) == "" {
		log.Warn("Could not update journal entry", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rating, err := strconv.Atoi(rrating)
	if err != nil {
		log.Warn("Could not update journal entry", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Updating journal entry in DB",
		"id", id,
		"title", title,
		"rating", rating,
	)

	updatedJournalEntry, err := c.persister.UpdateJournalEntry(r.Context(), int32(id), title, body, int32(rating), userData.Email)
	if err != nil {
		log.Warn("Could not update journal entry in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/journal/view?id=%v", updatedJournalEntry.ID), http.StatusFound)
}

func (c *Controller) HandleViewJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for view journal page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling view journal page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare view journal page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare view journal page", "err", errors.Join(errInvalidQueryParam, err))

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting journal entry for view",
		"id", id,
	)

	journalEntry, err := c.persister.GetJournalEntry(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get journal entry for view from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "journal_view.html", journalEntryData{
		pageData: pageData{
			userData: userData,

			Page:       journalEntry.Title,
			PrivacyURL: c.privacyURL,
			TosURL:     c.tosURL,
			ImprintURL: c.imprintURL,

			BackURL: "/journal",
		},
		Entry: journalEntry,
	}); err != nil {
		log.Warn("Could not render template for view journal page", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
