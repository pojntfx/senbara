package controllers

import (
	"fmt"
	"log"
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

func (b *Controller) HandleJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	journalEntries, err := b.persister.GetJournalEntries(r.Context(), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "journal.html", journalData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Journal"),
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Entries: journalEntries,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleAddJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := b.tpl.ExecuteTemplate(w, "journal_add.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Add a journal entry"),
		PrivacyURL: b.privacyURL,
		ImprintURL: b.imprintURL,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleCreateJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	body := r.FormValue("body")
	if strings.TrimSpace(body) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rrating := r.FormValue("rating")
	if strings.TrimSpace(rrating) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rating, err := strconv.Atoi(rrating)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := b.persister.CreateJournalEntry(r.Context(), title, body, int32(rating), userData.Email)
	if err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/journal/view?id=%v", id), http.StatusFound)
}

func (b *Controller) HandleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if err := b.persister.DeleteJournalEntry(r.Context(), int32(id), userData.Email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/journal", http.StatusFound)
}

func (b *Controller) HandleEditJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	journalEntry, err := b.persister.GetJournalEntry(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "journal_edit.html", journalEntryData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit journal entry"),
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Entry: journalEntry,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleUpdateJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	body := r.FormValue("body")
	if strings.TrimSpace(body) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rrating := r.FormValue("rating")
	if strings.TrimSpace(rrating) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rating, err := strconv.Atoi(rrating)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if err := b.persister.UpdateJournalEntry(r.Context(), int32(id), title, body, int32(rating), userData.Email); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/journal/view?id="+rid, http.StatusFound)
}

func (b *Controller) HandleViewJournal(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	journalEntry, err := b.persister.GetJournalEntry(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "journal_view.html", journalEntryData{
		pageData: pageData{
			userData: userData,

			Page:       journalEntry.Title,
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,

			BackURL: "/journal",
		},
		Entry: journalEntry,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
