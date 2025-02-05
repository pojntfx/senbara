package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

type activityData struct {
	pageData
	Entry models.GetActivityAndContactRow
}

func (c *Controller) HandleAddActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
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

	c.log.Debug("Getting contact for activity addition",
		"id", id,
		"email", userData.Email,
	)

	contact, err := c.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "activities_add.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Add an activity"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entry: contact,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleCreateActivity(w http.ResponseWriter, r *http.Request) {
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

	b.log.Debug("Creating activity",
		"contactID", contactID,
		"name", name,
		"date", date,
		"email", userData.Email,
	)

	if _, err := b.persister.CreateActivity(
		r.Context(),

		name,
		date,
		description,

		int32(contactID),
		userData.Email,
	); err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (b *Controller) HandleDeleteActivity(w http.ResponseWriter, r *http.Request) {
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

	b.log.Debug("Deleting activity",
		"id", id,
		"contactID", contactID,
		"email", userData.Email,
	)

	if err := b.persister.DeleteActivity(
		r.Context(),

		int32(id),

		userData.Email,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (b *Controller) HandleUpdateActivity(w http.ResponseWriter, r *http.Request) {
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

	b.log.Debug("Updating activity",
		"id", id,
		"name", name,
		"date", date,
		"email", userData.Email,
	)

	if err := b.persister.UpdateActivity(
		r.Context(),

		int32(id),

		userData.Email,

		name,
		date,
		description,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (b *Controller) HandleEditActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r, true)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	rid := r.URL.Query().Get("id")
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

	b.log.Debug("Getting activity and contact for edit",
		"id", id,
		"email", userData.Email,
	)

	activityAndContact, err := b.persister.GetActivityAndContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "activities_edit.html", activityData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit activity"),
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Entry: activityAndContact,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleViewActivity(w http.ResponseWriter, r *http.Request) {
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

	b.log.Debug("Getting activity and contact for view",
		"id", id,
		"contactID", contactID,
		"email", userData.Email,
	)

	activityAndContact, err := b.persister.GetActivityAndContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "activities_view.html", activityData{
		pageData: pageData{
			userData: userData,

			Page:       activityAndContact.Name,
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,

			BackURL: fmt.Sprintf("/contacts/view?id=%v", contactID),
		},
		Entry: activityAndContact,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
