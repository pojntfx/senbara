package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

func (b *Controller) HandleCreateActivity(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	name := r.FormValue("name")
	if strings.TrimSpace(name) == "" {
		log.Println(errInvalidForm)
		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)
		return
	}

	rdate := r.FormValue("date")
	if strings.TrimSpace(rdate) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	date, err := time.Parse("2006-01-02", rdate)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	id, err := b.persister.CreateActivity(
		r.Context(),

		name,
		date,
		description,

		int32(contactID),
		email,
	)
	if err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(models.GetActivitiesRow{
		ID: id,

		Name:        name,
		Date:        date,
		Description: description,
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleDeleteActivity(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if err := b.persister.DeleteActivity(
		r.Context(),

		int32(id),

		int32(contactID),
		email,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(id); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleUpdateActivity(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	name := r.FormValue("name")
	if strings.TrimSpace(name) == "" {
		log.Println(errInvalidForm)
		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)
		return
	}

	rdate := r.FormValue("date")
	if strings.TrimSpace(rdate) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	date, err := time.Parse("2006-01-02", rdate)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	if err := b.persister.UpdateActivity(
		r.Context(),

		int32(id),

		int32(contactID),
		email,

		name,
		date,
		description,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(models.GetActivitiesRow{
		ID: int32(id),

		Name:        name,
		Date:        date,
		Description: description,
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleViewActivity(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	rcontactID := r.URL.Query().Get("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	activityAndContact, err := b.persister.GetActivityAndContact(r.Context(), int32(id), int32(contactID), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(activityAndContact); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}
