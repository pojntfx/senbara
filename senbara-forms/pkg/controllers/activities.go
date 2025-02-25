package controllers

import (
	"errors"
	"fmt"
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
		c.log.Warn("Could not authorize user for add activity page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling add activity page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare add activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare add activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting contact to add activity to on add activity page from DB", "id", id)

	contact, err := c.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get contact to add activity from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

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
		log.Warn("Could not render template for adding an activity", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for create activity", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling create activity")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not create activity", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not create activity", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not create activity", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	name := r.FormValue("name")
	if strings.TrimSpace(name) == "" {
		log.Warn("Could not create activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rdate := r.FormValue("date")
	if strings.TrimSpace(rdate) == "" {
		log.Warn("Could not create activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	date, err := time.Parse("2006-01-02", rdate)
	if err != nil {
		log.Warn("Could not create activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	log.Debug("Creating activity in DB",
		"contactID", contactID,
		"name", name,
		"date", date,
		"description", description,
	)

	if _, err := c.persister.CreateActivity(
		r.Context(),

		name,
		date,
		description,

		int32(contactID),
		userData.Email,
	); err != nil {
		log.Warn("Could not create activity in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleDeleteActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for delete activity", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling delete activity")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not delete activity", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not delete activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not delete activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not delete activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not delete activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Deleting activity",
		"id", id,
		"contactID", contactID,
	)

	if _, err := c.persister.DeleteActivity(
		r.Context(),

		int32(id),

		userData.Email,
	); err != nil {
		log.Warn("Could not delete activity in DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleUpdateActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for update activity", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling update activity")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not update activity", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	name := r.FormValue("name")
	if strings.TrimSpace(name) == "" {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rdate := r.FormValue("date")
	if strings.TrimSpace(rdate) == "" {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	date, err := time.Parse("2006-01-02", rdate)
	if err != nil {
		log.Warn("Could not update activity", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	log.Debug("Updating activity",
		"id", id,
		"name", name,
		"date", date,
	)

	if _, err := c.persister.UpdateActivity(
		r.Context(),

		int32(id),

		userData.Email,

		name,
		date,
		description,
	); err != nil {
		log.Warn("Could not update activity in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleEditActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for edit activity page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling edit activity page")

	rid := r.URL.Query().Get("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare edit activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare edit activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting activity and contact for edit",
		"id", id,
	)

	activityAndContact, err := c.persister.GetActivityAndContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get activity and contact from DB for edit", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "activities_edit.html", activityData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit activity"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entry: activityAndContact,
	}); err != nil {
		log.Warn("Could not render template for editing an activity", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleViewActivity(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for view activity page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling view activity page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare view activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare view activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.URL.Query().Get("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not prepare view activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not prepare view activity page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting activity and contact for view",
		"id", id,
		"contactID", contactID,
	)

	activityAndContact, err := c.persister.GetActivityAndContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get activity and contact from DB for view", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "activities_view.html", activityData{
		pageData: pageData{
			userData: userData,

			Page:       activityAndContact.Name,
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,

			BackURL: fmt.Sprintf("/contacts/view?id=%v", contactID),
		},
		Entry: activityAndContact,
	}); err != nil {
		log.Warn("Could not render template for viewing an activity", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
