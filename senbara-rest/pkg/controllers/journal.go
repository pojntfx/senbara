package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (b *Controller) HandleJournal(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusUnauthorized)

		return
	}

	journalEntries, err := b.persister.GetJournalEntries(r.Context(), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(journalEntries); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleCreateJournal(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusUnauthorized)

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

	id, err := b.persister.CreateJournalEntry(r.Context(), title, body, int32(rating), email)
	if err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(models.JournalEntry{
		ID: id,

		Title:  title,
		Body:   body,
		Rating: int32(rating),
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusUnauthorized)

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

	if err := b.persister.DeleteJournalEntry(r.Context(), int32(id), email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(id); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleUpdateJournal(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusUnauthorized)

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

	if err := b.persister.UpdateJournalEntry(r.Context(), int32(id), title, body, int32(rating), email); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(models.JournalEntry{
		ID: int32(id),

		Title:  title,
		Body:   body,
		Rating: int32(rating),
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleViewJournal(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusUnauthorized)

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

	journalEntry, err := b.persister.GetJournalEntry(r.Context(), int32(id), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(journalEntry); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}
