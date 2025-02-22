package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

type contactsData struct {
	pageData
	Entries []models.Contact
}

type contactData struct {
	pageData
	Entry      models.Contact
	Debts      []models.GetDebtsRow
	Activities []models.GetActivitiesRow
}

func (c *Controller) HandleContacts(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for contacts page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling contacts page")

	contacts, err := c.persister.GetContacts(r.Context(), userData.Email)
	if err != nil {
		log.Warn("Could not get contacts from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "contacts.html", contactsData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Contacts"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entries: contacts,
	}); err != nil {
		log.Warn("Could not render contacts template", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleAddContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for add contact page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling add contact page")

	if err := c.tpl.ExecuteTemplate(w, "contacts_add.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Add a contact"),
		PrivacyURL: c.privacyURL,
		ImprintURL: c.imprintURL,
	}); err != nil {
		log.Warn("Could not render template for adding a contact", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for create contact", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling create contact")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not create contact", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	firstName := r.FormValue("first_name")
	if strings.TrimSpace(firstName) == "" {
		log.Warn("Could not create contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	lastName := r.FormValue("last_name")
	if strings.TrimSpace(lastName) == "" {
		log.Warn("Could not create contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		log.Warn("Could not create contact", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	nickname := r.FormValue("nickname")

	pronouns := r.FormValue("pronouns")
	if strings.TrimSpace(pronouns) == "" {
		log.Warn("Could not create contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Creating contact in DB",
		"firstName", firstName,
		"lastName", lastName,
		"nickname", nickname,
		"email", email,
		"pronouns", pronouns,
	)

	createdContact, err := c.persister.CreateContact(
		r.Context(),
		firstName,
		lastName,
		nickname,
		email,
		pronouns,
		userData.Email,
	)
	if err != nil {
		log.Warn("Could not create contact in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", createdContact.ID), http.StatusFound)
}

func (c *Controller) HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for delete contact", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling delete contact")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not delete contact", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not delete contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not delete contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Deleting contact from DB", "id", id)

	if _, err := c.persister.DeleteContact(r.Context(), int32(id), userData.Email); err != nil {
		log.Warn("Could not delete contact from DB", "err", errors.Join(errCouldNotDeleteFromDB, err))

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/contacts", http.StatusFound)
}

func (c *Controller) HandleViewContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for view contact page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling view contact page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare view contact page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare view contact page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting contact from DB", "id", id)

	contact, err := c.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	log.Debug("Getting debts for contact from DB", "id", id)

	debts, err := c.persister.GetDebts(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get debts from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	log.Debug("Getting activities for contact from DB", "id", id)

	activities, err := c.persister.GetActivities(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get activities from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "contacts_view.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       contact.FirstName + " " + contact.LastName,
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,

			BackURL: "/contacts",
		},
		Entry:      contact,
		Debts:      debts,
		Activities: activities,
	}); err != nil {
		log.Warn("Could not render template for viewing a contact", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for update contact", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling update contact")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not update contact", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not update contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not update contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	firstName := r.FormValue("first_name")
	if strings.TrimSpace(firstName) == "" {
		log.Warn("Could not update contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	lastName := r.FormValue("last_name")
	if strings.TrimSpace(lastName) == "" {
		log.Warn("Could not update contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		log.Warn("Could not update contact", "err", errors.Join(errInvalidForm, err))

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	nickname := r.FormValue("nickname")

	pronouns := r.FormValue("pronouns")
	if strings.TrimSpace(pronouns) == "" {
		log.Warn("Could not update contact", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rbirthday := r.FormValue("birthday")

	var birthday *time.Time
	if strings.TrimSpace(rbirthday) != "" {
		b, err := time.Parse("2006-01-02", rbirthday)
		if err != nil {
			log.Warn("Could not update contact", "err", errInvalidForm)

			http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

			return
		}

		birthday = &b
	}

	address := r.FormValue("address")

	notes := r.FormValue("notes")

	log.Debug("Updating contact in DB",
		"id", id,
		"firstName", firstName,
		"lastName", lastName,
		"nickname", nickname,
		"email", email,
		"pronouns", pronouns,
		"birthday", birthday,
		"address", address,
		"notes", notes,
	)

	updatedContact, err := c.persister.UpdateContact(
		r.Context(),
		int32(id),
		firstName,
		lastName,
		nickname,
		email,
		pronouns,
		userData.Email,
		birthday,
		address,
		notes,
	)
	if err != nil {
		log.Warn("Could not update contact in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/contacts/view?id="+string(updatedContact.ID), http.StatusFound)
}

func (c *Controller) HandleEditContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for edit contact page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling edit contact page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare edit contact page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare edit contact page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting contact for editing from DB", "id", id)

	contact, err := c.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "contacts_edit.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit contact"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entry: contact,
	}); err != nil {
		log.Warn("Could not render template for editing a contact", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
